package internal

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	"net/http"
)

type Controller struct {
	Sugar     *zap.SugaredLogger
	Mutations []Mutation
}

func (controller *Controller) HandleMutate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		controller.replyInternalServerError(w, "Error reading request body", err)
		return
	}

	admissionReview := v1.AdmissionReview{}
	err = json.Unmarshal(body, &admissionReview)
	if err != nil {
		controller.replyInternalServerError(w, "Error unmarshalling request", err)
		return
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		controller.replyInternalServerError(w, "Error unmarshalling request", err)
		return
	}

	patches, err := applyMutations(jsonMap, controller.Mutations)
	if err != nil {
		controller.replyInternalServerError(w, "Error applying mutations", err)
		return
	}

	patchesJSON, err := json.Marshal(patches)
	if err != nil {
		controller.replyInternalServerError(w, "Error marshalling patches", err)
		return
	}

	patchType := v1.PatchTypeJSONPatch
	//TODO Audit Annotations
	admissionResponse := v1.AdmissionResponse{
		UID:              admissionReview.Request.UID,
		Allowed:          true,
		PatchType:        &patchType,
		Patch:            patchesJSON,
	}

	admissionReview.Response = &admissionResponse
	admissionReview.Request = nil

	out, err := json.Marshal(admissionReview)
	if err != nil {
		controller.replyInternalServerError(w, "Error marshalling response", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(out))
}

// HandleHealth Handle health check requests
func (controller *Controller) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{status: 'OK'}"))

	if err != nil {
		controller.Sugar.Error(err)
	}
}

func (controller *Controller) replyInternalServerError(w http.ResponseWriter, msg string, err error) {
	controller.Sugar.Errorf("%s, %s", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s: %s", msg, err)
}
