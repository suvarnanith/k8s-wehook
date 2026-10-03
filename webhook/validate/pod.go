package validate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// validatePod holds all validation rules specific to Pod objects.
// Add new Pod rules here; create a new file for other kinds (e.g. service.go).
func validatePod(
	request *admissionv1.AdmissionRequest,
	response *admissionv1.AdmissionResponse,
) {
	var pod corev1.Pod

	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		log.Printf("BLOCK pod: failed to decode Pod: %v", err)

		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Message: fmt.Sprintf("failed to decode Pod: %v", err),
			Reason:  metav1.StatusReasonInvalid,
			Code:    http.StatusBadRequest,
		}
		return
	}

	namespace := request.Namespace

	exemptNamespace :=
		strings.HasPrefix(namespace, "kube-")

	if exemptNamespace {
		log.Printf(
			"ALLOW pod %s/%s: namespace is exempt",
			namespace,
			pod.Name,
		)
		return
	}

	// Validate normal containers.
	for _, container := range pod.Spec.Containers {
		if len(container.Resources.Requests) == 0 ||
			len(container.Resources.Limits) == 0 {

			message := fmt.Sprintf(
				"container %q must specify resource requests and limits",
				container.Name,
			)

			log.Printf(
				"BLOCK pod %s/%s: %s",
				namespace,
				pod.Name,
				message,
			)

			response.Allowed = false
			response.Result = &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: message,
				Reason:  metav1.StatusReasonForbidden,
				Code:    http.StatusForbidden,
			}
			return
		}
	}

	// Validate init containers too.
	for _, container := range pod.Spec.InitContainers {
		if len(container.Resources.Requests) == 0 ||
			len(container.Resources.Limits) == 0 {

			message := fmt.Sprintf(
				"init container %q must specify resource requests and limits",
				container.Name,
			)

			log.Printf(
				"BLOCK pod %s/%s: %s",
				namespace,
				pod.Name,
				message,
			)

			response.Allowed = false
			response.Result = &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: message,
				Reason:  metav1.StatusReasonForbidden,
				Code:    http.StatusForbidden,
			}
			return
		}
	}

	log.Printf("ALLOW pod %s/%s", namespace, pod.Name)
}
