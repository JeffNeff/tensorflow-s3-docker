/*
Copyright (c) 2021 TriggerMesh Inc.

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

package main

type B64ResponseEvent struct {
	B64 string `json:"b64"`
	URL string `json:"url"`
}

type TensorflowRequest struct {
	Instances []struct {
		B64 string `json:"b64"`
	} `json:"instances"`
}

type TensorflowResponse struct {
	Predictions []struct {
		// DetectionClasses          []int       `json:"detection_classes"`
		NumDetections             float64     `json:"num_detections"`
		DetectionBoxes            [][]float64 `json:"detection_boxes"`
		RawDetectionBoxes         [][]float64 `json:"raw_detection_boxes"`
		DetectionScores           []float64   `json:"detection_scores"`
		RawDetectionScores        [][]float64 `json:"raw_detection_scores"`
		DetectionMulticlassScores [][]float64 `json:"detection_multiclass_scores"`
	} `json:"predictions"`
}
