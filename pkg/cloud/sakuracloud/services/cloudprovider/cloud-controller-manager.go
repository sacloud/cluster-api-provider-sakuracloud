/*
Copyright 2019 Kazumichi Yamamoto.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudprovider

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ccmCredentialSecretName = "cloud-controller-manager-credential"

// NOTE: https://github.com/sacloud/sakura-cloud-controller-manager/blob/master/manifests/cloud-controller-manager.yaml

// CloudControllerManagerServiceAccount returns the ServiceAccount used for the cloud-controller-manager
func CloudControllerManagerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloud-controller-manager",
			Namespace: "kube-system",
		},
	}
}

func CloudControllerManagerCredential(token, secret string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ccmCredentialSecretName,
			Namespace: "kube-system",
		},
		StringData: map[string]string{
			"token":  token,
			"secret": secret,
		},
		Type: corev1.SecretTypeOpaque,
	}
}

// CloudControllerManagerDeployment returns the Deployment which runs the cloud-controller-manager
func CloudControllerManagerDeployment(image, zone, clusterID string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sakura-cloud-controller-manager",
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "sakura-cloud-controller-manager",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "sakura-cloud-controller-manager",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "sakura-cloud-controller-manager",
					},
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:    "node.cloudprovider.kubernetes.io/uninitialized",
							Value:  "true",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node.kubernetes.io/not-ready",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					ServiceAccountName: "cloud-controller-manager",
					Containers: []corev1.Container{
						{
							Name:  "sakura-cloud-controller-manager",
							Image: image,
							Args: []string{
								"/sakura-cloud-controller-manager",
								"--cloud-provider=sakuracloud",
								"--leader-elect=true",
								"--use-service-account-credentials=true",
								"--allocate-node-cidrs=false",
								"--configure-cloud-routes=false",
							},
							Env: []corev1.EnvVar{
								{
									Name: "SAKURACLOUD_ACCESS_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "token",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ccmCredentialSecretName,
											},
										},
									},
								},
								{
									Name: "SAKURACLOUD_ACCESS_TOKEN_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "secret",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ccmCredentialSecretName,
											},
										},
									},
								},
								{
									Name:  "SAKURACLOUD_ZONE",
									Value: zone,
								},
								{
									Name:  "SAKURACLOUD_CLUSTER_ID",
									Value: clusterID,
								},
								// We don't use sakura cloud native load balancer.
								{
									Name:  "SAKURACLOUD_DISABLE_LOAD_BALANCER",
									Value: "1",
								},
							},
						},
					},
					DNSPolicy:   corev1.DNSDefault,
					HostNetwork: true,
				},
			},
			RevisionHistoryLimit: int32ptr(2),
		},
	}
}

// CloudControllerManagerClusterRole returns the ClusterRole systemLcloud-controller-manager used by the cloud-controller-manager
func CloudControllerManagerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"cofigmaps"},
				Verbs:     []string{"create", "get", "list", "watch", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"cofigmaps/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

// CloudControllerManagerRoleBinding binds the extension-apiserver-authentication-reader to the cloud-controller-manager
func CloudControllerManagerRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "servicecatalog.k8s.io:apiserver-authentication-reader",
			Namespace: "kube-system",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "extension-apiserver-authentication-reader",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  "",
				Kind:      "ServiceAccount",
				Name:      "cloud-controller-manager",
				Namespace: "kube-system",
			},
			{
				APIGroup: "",
				Kind:     "User",
				Name:     "cloud-controller-manager",
			},
		},
	}
}

// CloudControllerManagerClusterRoleBinding binds the system:cloud-controller-manager cluster role to the cloud-controller-manager
func CloudControllerManagerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:cloud-controller-manager",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "cloud-controller-manager",
				Namespace: "kube-system",
			},
			{
				Kind: "User",
				Name: "cloud-controller-manager",
			},
		},
	}
}

func int32ptr(i int) *int32 {
	ptr := int32(i)
	return &ptr
}
