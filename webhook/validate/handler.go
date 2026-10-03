package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handler receives the AdmissionReview and dispatches to the validator for
// the requested kind. Add new "else if" branches (e.g. Service, Deployment)
// here, calling into a dedicated file such as service.go.
func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got request...")

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

		if review.Request.Kind.Group == "" &&
			review.Request.Kind.Version == "v1" &&
			review.Request.Kind.Kind == "Pod" {

			validatePod(review.Request, response)
		}
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
		"Admission response sent: uid=%s allowed=%t message=%q",
		response.UID,
		response.Allowed,
		statusMessage(response.Result),
	)
}

func statusMessage(status *metav1.Status) string {
	if status == nil {
		return ""
	}

	return status.Message
}
