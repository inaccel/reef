package internal

import (
	"encoding/json"
	"net/http"

	"github.com/inaccel/reef/pkg/jsonpatch"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Rules = []admissionregistrationv1.RuleWithOperations{
	{
		Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
		Rule: admissionregistrationv1.Rule{
			APIGroups:   []string{""},
			APIVersions: []string{"v1"},
			Resources:   []string{"pods"},
		},
	},
}

type Func func(corev1.Pod) (corev1.Pod, error)

type Webhook struct {
	f Func
}

func Handle(pattern string) http.Handler {
	handler := http.NewServeMux()
	handler.Handle(pattern, &Webhook{Mutate})
	return handler
}

func (webhook Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"request": review.Request,
	}).Debug("review")

	review.Response = &admissionv1.AdmissionResponse{
		UID: review.Request.UID,
	}

	var before corev1.Pod
	if err := json.Unmarshal(review.Request.Object.Raw, &before); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	review.Request = nil

	if after, err := webhook.f(before); err != nil {
		review.Response.Allowed = false

		review.Response.Result = &metav1.Status{
			Message: err.Error(),
		}
	} else {
		review.Response.Allowed = true

		patch, err := jsonpatch.Diff(before, after)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		patchType := admissionv1.PatchTypeJSONPatch

		review.Response.Patch = patch
		review.Response.PatchType = &patchType
	}

	logrus.WithFields(logrus.Fields{
		"response": review.Response,
	}).Debug("review")

	if err := json.NewEncoder(w).Encode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
