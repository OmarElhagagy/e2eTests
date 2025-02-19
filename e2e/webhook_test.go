//nolint:all
package e2e_test

import (
	"context"
	"fmt"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argoprojv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mutating Webhooks", func() {
	const (
		testNamespace      = "argocd"
		defaultTimeout     = time.Second * 10
		defaultTestTenant  = "test-tenat"
		defaultAppName     = "test-app"
		defaultAppSetName  = "test-appset"
		defaultRepoURL     = "https://github.com/argoproj/argocd-example-apps"
		defaultPath        = "guestbook"
		defaultDestination = "in-cluster"
	)

	var (
		ctx        context.Context
		tenant     *capsulev1beta2.Tenant
		translator *v1alpha1.ArgoTranslator
	)

	BeforeEach(func() {
		ctx = context.Background()
		// test tenat for our test
		tenant = &capsulev1beta2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:   defaultTestTenant,
				Labels: e2eLabels("webhook-test"),
			},
			Spec: capsulev1beta2.TenantSpec{
				Owners: capsulev1beta2.OwnerListSpec{
					{
						Name: "webhook-test-owner",
						Kind: "User",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, tenant)).To(Succeed())

		// Translator for our test
		translator = &v1alpha1.ArgoTranslator{
			ObjectMeta: metav1.ObjectMeta{
				Name:   defaultTestTenant,
				Labels: e2eLabels("webhook-test"),
			},
			Spec: v1alpha1.ArgoTranslatorSpec{
				TenantName: defaultTestTenant,
				ArgoCD: v1alpha1.ArgoTranslatorSpec{
					Destination: defaultDestination,
					Namespace:   testNamespace,
					Project:     defaultTestTenant,
				},
			},
		}
		Expect(k8sClient.Create(ctx, translator)).To(Succeed())

		// ArgoCD project
		project := &argocdv1alpha1.AppProject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultTestTenant,
				Namespace: testNamespace,
				Labels:    e2eLabels("webhook-test"),
			},
			Spec: argocdv1alpha1.AppProjectSpec{
				SourceRepos: []string{"*"},
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "*",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, project)).To(Succeed())
	})
	AfterEach(func() {
		selector := e2eSelector("webhook-test")
		Expect(CleanTenants(selector)).To(Succeed())
		Expect(CleanTranslators(selector)).To(Succeed())
		Expect(CleanAppProjects(selector, testNamespace)).To(Succeed())
		Expect(cleanApplications(selector, testNamespace)).To(Succeed())
		Expect(cleanApplicationSets(selector, testNamespace)).To(Succeed())
	})

	Context("Application Mutating Webhook", func() {
		When("creating an application with webhook enabled", func() {
			It("should mutate the application correctly", func() {
				// update ArgoAddon to enable webhook for application
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "default"}, argoAddon)).To(Succeed())
				argoAddon.Spec.ApplicationWebhook = true
				Expect(k8sClient.Update(ctx, argoAddon)).To(Succeed())

				// create an application
				app := &argocdv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaultAppName,
						Namespace: testNamespace,
						Labels: map[string]string{
							"capsule.clastix.io/tenant": defaultTestTenant,
							e2eLabel:                    "true",
							suiteLabel:                  "webhook-test",
						},
					},
					Spec: argocdv1alpha1.ApplicationSpec{
						Project: "default", // this should be mutated to tenant name
						Source: &argocdv1alpha1.ApplicationSource{
							RepoURL:        defaultRepoURL,
							Path:           defaultPath,
							TargetRevision: "HEAD",
						},
						Destination: argocdv1alpha1.ApplicationDestination{
							Server:    "https://kubernetes.default.svc",
							Namespace: "default",
						},
					},
				}

				Expect(k8sClient.Create(ctx, app)).To(Succeed())

				// verify that the app is mutated
				Eventually(func() bool {
					updatedApp := &argocdv1alpha1.Application{}
					err := k8sClient.Get(ctx, client.ObjectKey{Name: defaultAppName, Namespace: testNamespace}
						,updatedApp)
					if err != nil {
						return false
					}
					// Project should be mutated to tenant name
					if updatedApp.Spex.Project != defaultTestTenant {
						return false
					}
					// Check for additional labels/annotations added by webhook
					return true
				}, defaultTdefaultTimeout, defadefaultPollInterval).Should(BeTBeTrue())
			})
		})
	})
	
	Context("ApplicationSet Mutating Webhook", func() {
		When("Creating an ApplicationSet with webhook enabled", func() {
			It("Should mutate the ApplicatonSet correctly", func() {
				// Update ArgoAddon to enable webhook for ApplicationSets
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "default"}, argoargoAddon)).To(Succeed())
				argoAddon.Spec.ApplicationSetWebhook = true
				Expect(k8sClient.Update(ctx, argoAddon)).To(Succeed())

				// create an ApplicationSet
				appSet := &argocdv1alpha1.ApplicationSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultAppSetName,
						NameSpace: testNameSpace,
						Labels: map[string]string{
							"capsule.clastix.io/tenant": defaultTestTenant,
							e2eLabel: "true",
							suiteLabel: "webhook-test",
						},
					},
					Spec: argocdv1alpha1.ApplicationSetSpec{
						Generators: []argocdv1alpha1.ApplicationSetGenerator{
							{
								List: &argocdv1alpha1.ListGenerator{
									Elements: []map[string]interface{}{
										"cluster": "in-cluster",
										"url": "https://kubernetes.default.svc",
										"revision": "HEAD",
									},
								},
							},
						},
					},
					Template:
				}
			})
		})
	})

})
