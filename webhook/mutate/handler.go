package mutate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handler receives the AdmissionReview and adds the environment=default
// label when it's missing. To add a new mutation, create a dedicated file
// in this package and wire it into mutateObject below.
func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got mutate request...")

	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview

	if err := json.Unmarshal(body, &review); err != nil {
		log.Printf("Failed to decode AdmissionReview: %v", err)
		http.Error(w, "failed to decode AdmissionReview", http.StatusBadRequest)
		return
	}

	response := &admissionv1.AdmissionResponse{
		Allowed: true,
	}

	if review.Request == nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Message: "AdmissionReview request is missing",
			Reason:  metav1.StatusReasonInvalid,
			Code:    http.StatusBadRequest,
		}
	} else {
		response.UID = review.Request.UID

		mutateObject(review.Request, response)
	}

	responseReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: response,
	}

	payload, err := json.Marshal(responseReview)
	if err != nil {
		log.Printf("Failed to marshal AdmissionReview response: %v", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(payload); err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

	log.Printf(
		"Mutation response sent: uid=%s allowed=%t patched=%t",
		response.UID,
		response.Allowed,
		len(response.Patch) > 0,
	)
}

// mutateObject adds the environment=default label to the object's metadata
// when it doesn't already declare an "environment" label.
func mutateObject(request *admissionv1.AdmissionRequest, response *admissionv1.AdmissionResponse) {
	kind := request.Kind.Kind

	patch, err := buildEnvironmentLabelPatch(request.Object.Raw)
	if err != nil {
		log.Printf("BLOCK %s %s/%s: %v", kind, request.Namespace, request.Name, err)

		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
			Reason:  metav1.StatusReasonInvalid,
			Code:    http.StatusBadRequest,
		}
		return
	}

	if patch == nil {
		log.Printf("ALLOW %s %s/%s: %s label already present", kind, request.Namespace, request.Name, environmentLabelKey)
		return
	}

	patchType := admissionv1.PatchTypeJSONPatch
	response.Patch = patch
	response.PatchType = &patchType

	log.Printf(
		"MUTATE %s %s/%s: added label %s=%s",
		kind,
		request.Namespace,
		request.Name,
		environmentLabelKey,
		environmentLabelDefaultValue,
	)
}
