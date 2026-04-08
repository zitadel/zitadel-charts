package main

import (
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	. "github.com/dave/jennifer/jen"
)

// assertgenPkgs is the set of K8s API packages scanned by assertgen.
// Resources whose types come from these packages have assertion structs
// and are included in the AssertPartial type switch.
var assertgenPkgs = map[string]bool{
	"k8s.io/api/apps/v1":        true,
	"k8s.io/api/autoscaling/v2": true,
	"k8s.io/api/batch/v1":       true,
	"k8s.io/api/core/v1":        true,
	"k8s.io/api/networking/v1":   true,
	"k8s.io/api/policy/v1":       true,
	"k8s.io/api/rbac/v1":         true,
	"k8s.io/api/storage/v1":      true,
}

type resource struct {
	Name         string // possibly disambiguated, e.g. "AdmissionregistrationV1alpha1MutatingAdmissionPolicy"
	TypeName     string // original Go type name in the package, e.g. "MutatingAdmissionPolicy"
	TypePkgPath  string // e.g. "k8s.io/api/apps/v1"
	GroupMethod  string // e.g. "AppsV1"
	Plural       string // e.g. "Deployments" (the method name on the group interface)
	Namespaced   bool
	HasAssertion bool   // whether an assertion struct exists for this type
}

func (r resource) AssertName() string {
	return r.Name + "Assertion"
}

func main() {
	outFlag := flag.String("out", "", "output file path")
	pkgFlag := flag.String("package", "support", "package name for the generated file")
	flag.Parse()

	if *outFlag == "" {
		log.Fatal("-out flag is required")
	}

	// Load the clientset package to introspect kubernetes.Interface
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, "k8s.io/client-go/kubernetes")
	if err != nil {
		log.Fatalf("loading packages: %v", err)
	}
	if len(pkgs) == 0 {
		log.Fatal("no packages loaded")
	}
	if len(pkgs[0].Errors) > 0 {
		for _, e := range pkgs[0].Errors {
			log.Printf("package error: %v", e)
		}
		log.Fatal("package had errors")
	}

	// Find kubernetes.Interface
	clientsetPkg := pkgs[0]
	ifaceObj := clientsetPkg.Types.Scope().Lookup("Interface")
	if ifaceObj == nil {
		log.Fatal("kubernetes.Interface not found")
	}
	ifaceType, ok := ifaceObj.Type().Underlying().(*types.Interface)
	if !ok {
		log.Fatal("kubernetes.Interface is not an interface")
	}

	// Discover resources by walking the Interface methods
	var resources []resource
	for i := 0; i < ifaceType.NumMethods(); i++ {
		groupMethod := ifaceType.Method(i)
		if groupMethod.Name() == "Discovery" {
			continue
		}

		groupSig := groupMethod.Type().(*types.Signature)
		if groupSig.Results().Len() != 1 {
			continue
		}

		// Get the group interface (e.g., appsv1.AppsV1Interface)
		groupRetType := groupSig.Results().At(0).Type()
		groupIface, ok := groupRetType.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		// Get the complete method set of the group interface
		mset := types.NewMethodSet(groupRetType)
		for j := 0; j < mset.Len(); j++ {
			sel := mset.At(j)
			fn, ok := sel.Obj().(*types.Func)
			if !ok {
				continue
			}
			_ = groupIface // used above

			// Skip RESTClient and other non-resource methods
			if fn.Name() == "RESTClient" {
				continue
			}

			fnSig, ok := fn.Type().(*types.Signature)
			if !ok {
				continue
			}

			// The getter method either takes (namespace string) for namespaced
			// resources or no params for cluster-scoped resources.
			// It returns a single resource interface.
			if fnSig.Results().Len() != 1 {
				continue
			}

			namespaced := fnSig.Params().Len() == 1

			// Get the resource interface return type
			resRetType := fnSig.Results().At(0).Type()

			// Find the Get() method on the resource interface
			resMset := types.NewMethodSet(resRetType)
			var getMethod *types.Func
			for k := 0; k < resMset.Len(); k++ {
				m := resMset.At(k)
				if m.Obj().Name() == "Get" {
					getMethod = m.Obj().(*types.Func)
					break
				}
			}
			if getMethod == nil {
				continue
			}

			getSig := getMethod.Type().(*types.Signature)
			if getSig.Results().Len() < 1 {
				continue
			}

			// Extract the K8s type from the Get() return value (it's a pointer type)
			getRetType := getSig.Results().At(0).Type()
			ptr, ok := getRetType.(*types.Pointer)
			if !ok {
				continue
			}
			named, ok := ptr.Elem().(*types.Named)
			if !ok {
				continue
			}

			typePkg := named.Obj().Pkg()
			if typePkg == nil {
				continue
			}

			typePkgPath := typePkg.Path()
			typeName := named.Obj().Name()

			// Check if this type has an assertion struct.
			// assertgen skips types ending in "List" or "Status".
			hasAssertion := assertgenPkgs[typePkgPath] &&
				!strings.HasSuffix(typeName, "List") &&
				!strings.HasSuffix(typeName, "Status")

			resources = append(resources, resource{
				Name:         typeName,
				TypeName:     typeName,
				TypePkgPath:  typePkgPath,
				GroupMethod:  groupMethod.Name(),
				Plural:       fn.Name(),
				Namespaced:   namespaced,
				HasAssertion: hasAssertion,
			})
		}
	}

	// Resolve method-name collisions instead of deduplicating by type name.
	// The same logical K8s type (e.g. HorizontalPodAutoscaler) can appear in
	// multiple API group versions (v1, v2, v1beta1). We still want a single
	// "short" method name (e.g. GetHorizontalPodAutoscaler) for the best API
	// version, but we must not drop distinct resources that happen to share a
	// type name across different API groups (e.g. core/v1 and events.k8s.io/v1 Event).
	//
	// To achieve this, we:
	//   * keep all discovered resources; and
	//   * for each set of resources that share the same type Name, we pick the
	//     best one (by resourcePriority) to keep the original Name, and we
	//     disambiguate the others by prefixing their Name with GroupMethod.
	// This preserves existing short method names for the preferred version while
	// ensuring that all distinct resources get generated methods with unique names.
	nameToIdxs := make(map[string][]int)
	for i := range resources {
		r := &resources[i]
		nameToIdxs[r.Name] = append(nameToIdxs[r.Name], i)
	}
	for _, idxs := range nameToIdxs {
		if len(idxs) <= 1 {
			continue
		}
		// Choose the best resource to keep the original (short) Name.
		bestIdx := idxs[0]
		for _, idx := range idxs[1:] {
			if resourcePriority(resources[idx]) > resourcePriority(resources[bestIdx]) {
				bestIdx = idx
			}
		}
		// Disambiguate all non-best resources by prefixing with GroupMethod.
		for _, idx := range idxs {
			if idx == bestIdx {
				continue
			}
			r := &resources[idx]
			// Only change the name if it would actually collide; this is defensive
			// in case future code mutates r.Name earlier.
			r.Name = r.GroupMethod + r.Name
		}
	}

	// Sort for deterministic output
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].GroupMethod != resources[j].GroupMethod {
			return resources[i].GroupMethod < resources[j].GroupMethod
		}
		return resources[i].Name < resources[j].Name
	})

	// Build the jennifer file
	f := NewFile(*pkgFlag)
	f.HeaderComment("Code generated by supportgen. DO NOT EDIT.")

	// Emit getter methods for each resource
	for _, r := range resources {
		// Get<Resource>(t, name) *pkg.Type
		f.Comment(fmt.Sprintf("Get%s fetches a %s by name, failing the test on error.", r.Name, r.Name))
		f.Func().Params(Id("env").Op("*").Id("Env")).Id("Get"+r.Name).Params(
			Id("t").Op("*").Qual("testing", "T"),
			Id("name").String(),
		).Op("*").Qual(r.TypePkgPath, r.TypeName).Block(
			Id("t").Dot("Helper").Call(),
			getClientCall(r, false),
			Qual("github.com/stretchr/testify/require", "NoError").Call(
				Id("t"), Id("err"), Lit(fmt.Sprintf("failed to get %s %%s", r.Name)), Id("name"),
			),
			Return(Id("obj")),
		)
		f.Line()

		// Get<Resource>E(t, name) (*pkg.Type, error)
		f.Comment(fmt.Sprintf("Get%sE fetches a %s by name, returning the error for non-existence checks.", r.Name, r.Name))
		f.Func().Params(Id("env").Op("*").Id("Env")).Id("Get"+r.Name+"E").Params(
			Id("t").Op("*").Qual("testing", "T"),
			Id("name").String(),
		).Parens(List(
			Op("*").Qual(r.TypePkgPath, r.TypeName),
			Error(),
		)).Block(
			Id("t").Dot("Helper").Call(),
			Return(getClientExpr(r)),
		)
		f.Line()
	}

	// Collect resources with assertion types for the AssertPartial switch
	var assertResources []resource
	for _, r := range resources {
		if r.HasAssertion {
			assertResources = append(assertResources, r)
		}
	}

	// Emit AssertPartial method
	var switchCases []Code
	for _, r := range assertResources {
		switchCases = append(switchCases,
			Case(
				Qual("github.com/zitadel/zitadel-charts/test/assert", r.AssertName()),
			).Block(
				Qual("github.com/zitadel/zitadel-charts/test/assert", "AssertPartial").Call(
					Id("t"),
					Id("env").Dot("Get"+r.Name).Call(Id("t"), Id("name")),
					Id("a"),
					Id("name"),
				),
			),
			Case(
				Op("*").Qual("github.com/zitadel/zitadel-charts/test/assert", r.AssertName()),
			).Block(
				Qual("github.com/zitadel/zitadel-charts/test/assert", "AssertPartial").Call(
					Id("t"),
					Id("env").Dot("Get"+r.Name).Call(Id("t"), Id("name")),
					Op("*").Id("a"),
					Id("name"),
				),
			),
		)
	}
	switchCases = append(switchCases,
		Default().Block(
			Id("env").Dot("assertPartialFallback").Call(Id("t"), Id("name"), Id("assertion")),
		),
	)

	f.Comment("AssertPartial fetches the K8s resource implied by the assertion type and")
	f.Comment("performs a partial assertion. The resource type is inferred from the")
	f.Comment("concrete assertion struct via a type switch.")
	f.Func().Params(Id("env").Op("*").Id("Env")).Id("AssertPartial").Params(
		Id("t").Op("*").Qual("testing", "T"),
		Id("name").String(),
		Id("assertion").Qual("github.com/zitadel/zitadel-charts/test/assert", "Assertable"),
	).Block(
		Id("t").Dot("Helper").Call(),
		Switch(Id("a").Op(":=").Id("assertion").Assert(Type())).Block(switchCases...),
	)
	f.Line()

	// Emit AssertNone method — asserts that a resource does NOT exist
	var noneSwitchCases []Code
	for _, r := range assertResources {
		noneSwitchCases = append(noneSwitchCases,
			Case(
				Qual("github.com/zitadel/zitadel-charts/test/assert", r.AssertName()),
				Op("*").Qual("github.com/zitadel/zitadel-charts/test/assert", r.AssertName()),
			).Block(
				List(Id("_"), Id("err")).Op(":=").Id("env").Dot("Get"+r.Name+"E").Call(Id("t"), Id("name")),
				Qual("github.com/stretchr/testify/require", "True").Call(
					Id("t"),
					Qual("k8s.io/apimachinery/pkg/api/errors", "IsNotFound").Call(Id("err")),
					Lit(fmt.Sprintf("%s %%q should not exist (err: %%v)", r.Name)),
					Id("name"),
					Id("err"),
				),
			),
		)
	}
	noneSwitchCases = append(noneSwitchCases,
		Default().Block(
			Id("env").Dot("assertNoneFallback").Call(Id("t"), Id("name"), Id("assertion")),
		),
	)

	f.Comment("AssertNone asserts that the K8s resource implied by the assertion type")
	f.Comment("does not exist. The resource type is inferred from the concrete assertion")
	f.Comment("struct via a type switch.")
	f.Func().Params(Id("env").Op("*").Id("Env")).Id("AssertNone").Params(
		Id("t").Op("*").Qual("testing", "T"),
		Id("name").String(),
		Id("assertion").Qual("github.com/zitadel/zitadel-charts/test/assert", "Assertable"),
	).Block(
		Id("t").Dot("Helper").Call(),
		Switch(Id("assertion").Assert(Type())).Block(noneSwitchCases...),
	)

	// Render to string
	var buf strings.Builder
	if err := f.Render(&buf); err != nil {
		log.Fatalf("rendering jennifer: %v", err)
	}

	// Format
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		_ = os.WriteFile(*outFlag+".raw", []byte(buf.String()), 0644)
		log.Fatalf("gofmt: %v (raw output written to %s.raw)", err, *outFlag)
	}

	if err := os.WriteFile(*outFlag, formatted, 0644); err != nil {
		log.Fatalf("writing output: %v", err)
	}

	fmt.Printf("Generated %s with %d resources (%d with assertions, %d bytes)\n",
		*outFlag, len(resources), len(assertResources), len(formatted))
}

// resourcePriority computes a priority score for deduplication.
// Higher is better. Prefers: assertion support > stable > beta > alpha > higher version.
func resourcePriority(r resource) int {
	p := 0

	// Strongly prefer resources with assertion structs
	if r.HasAssertion {
		p += 1000
	}

	// Prefer stable over beta over alpha
	pkgPath := r.TypePkgPath
	if strings.Contains(pkgPath, "alpha") {
		p += 0
	} else if strings.Contains(pkgPath, "beta") {
		p += 100
	} else {
		p += 200
	}

	// Higher version number wins as tiebreaker
	p += extractVersion(pkgPath)
	return p
}

// extractVersion extracts the version number from a package path.
// e.g. "k8s.io/api/autoscaling/v2" -> 2, "k8s.io/api/apps/v1" -> 1.
func extractVersion(pkgPath string) int {
	parts := strings.Split(pkgPath, "/")
	last := parts[len(parts)-1]
	// Strip non-digit prefix (e.g. "v1beta2" -> look for digits after 'v')
	v := 0
	for _, c := range last {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		} else if v > 0 {
			break // stop at first non-digit after digits
		}
	}
	return v
}

// getClientCall generates: obj, err := env.Client.<Group>().<Plural>([namespace]).Get(ctx, name, opts)
func getClientCall(r resource, _ bool) *Statement {
	return List(Id("obj"), Id("err")).Op(":=").Add(getClientExpr(r))
}

// getClientExpr generates: env.Client.<Group>().<Plural>([namespace]).Get(ctx, name, opts)
func getClientExpr(r resource) *Statement {
	chain := Id("env").Dot("Client").Dot(r.GroupMethod).Call()
	if r.Namespaced {
		chain = chain.Dot(r.Plural).Call(Id("env").Dot("Kube").Dot("Namespace"))
	} else {
		chain = chain.Dot(r.Plural).Call()
	}
	chain = chain.Dot("Get").Call(
		Id("env").Dot("Ctx"),
		Id("name"),
		Qual("k8s.io/apimachinery/pkg/apis/meta/v1", "GetOptions").Values(),
	)
	return chain
}
