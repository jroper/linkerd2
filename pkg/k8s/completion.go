package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Complete(ctx context.Context, k8sAPI *KubernetesAPI, namespace string, args []string, toComplete string) ([]string, error) {
	autoCmpRegexp := regexp.MustCompile(fmt.Sprintf("^%s.*", toComplete))
	if len(args) == 0 && toComplete == "" {
		return StatAllResourceTypes, nil
	}

	if len(args) == 0 && toComplete != "" {
		targets := []string{}
		for _, t := range StatAllResourceTypes {
			if autoCmpRegexp.MatchString(t) {
				targets = append(targets, t)
			}
		}
		return targets, nil
	}

	if len(args) == 1 {
		resType, err := PluralResourceNameFromFriendlyName(args[0])
		if err != nil {
			return nil, fmt.Errorf("%s not a valid resource name", args)
		}

		apiResourceList, err := k8sAPI.Discovery().ServerPreferredNamespacedResources()
		if err != nil {
			return nil, err
		}

		var gvr *schema.GroupVersionResource
		for _, res := range apiResourceList {
			for _, r := range res.APIResources {
				if r.Name == resType {
					gv := strings.Split(res.GroupVersion, "/")

					if len(gv) == 1 && gv[0] == "v1" {
						gvr = &schema.GroupVersionResource{
							Version:  gv[0],
							Resource: r.Name,
						}
						break
					}

					if len(gv) != 2 {
						return nil, fmt.Errorf("could not find the requested resource")
					}

					gvr = &schema.GroupVersionResource{
						Group:    gv[0],
						Version:  gv[1],
						Resource: r.Name,
					}
					break
				}

				if gvr != nil {
					break
				}
			}
		}

		uList, err := k8sAPI.DynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		targets := []string{}
		for _, u := range uList.Items {
			name := u.GetName()
			if autoCmpRegexp.MatchString(name) {
				targets = append(targets, name)
			}
		}

		return targets, nil
	}

	return []string{}, nil
}

func FlagComplete(ctx context.Context, k8sAPI *KubernetesAPI, resource string, toComplete string) ([]string, error) {
	autoCmpRegexp := regexp.MustCompile(fmt.Sprintf("^%s.*", toComplete))

	apiResourceList, err := k8sAPI.Discovery().ServerPreferredResources()
	if err != nil {
		fmt.Printf("discovery %s", err.Error())
		return nil, err
	}

	var gvr *schema.GroupVersionResource
	for _, res := range apiResourceList {
		for _, r := range res.APIResources {
			if r.Name == resource {
				gv := strings.Split(res.GroupVersion, "/")

				if len(gv) == 1 && gv[0] == "v1" {
					gvr = &schema.GroupVersionResource{
						Version:  gv[0],
						Resource: r.Name,
					}
					break
				}

				if len(gv) != 2 {
					fmt.Printf("not found resource %s", err.Error())
					return nil, fmt.Errorf("could not find the requested resource")
				}

				gvr = &schema.GroupVersionResource{
					Group:    gv[0],
					Version:  gv[1],
					Resource: r.Name,
				}
				break
			}

			if gvr != nil {
				break
			}
		}
	}

	uList, err := k8sAPI.DynamicClient.Resource(*gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("list failed %s %+v", err.Error(), *gvr)
		return nil, err
	}

	targets := []string{}
	for _, u := range uList.Items {
		name := u.GetName()
		if autoCmpRegexp.MatchString(name) {
			targets = append(targets, name)
		}
	}

	return targets, nil
}
