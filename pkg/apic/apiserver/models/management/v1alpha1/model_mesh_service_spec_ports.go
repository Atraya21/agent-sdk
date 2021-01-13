/*
 * API Server specification.
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: SNAPSHOT
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1alpha1

// MeshServiceSpecPorts struct for MeshServiceSpecPorts
type MeshServiceSpecPorts struct {
	Name      string                     `json:"name,omitempty"`
	Port      int32                      `json:"port,omitempty"`
	Endpoints []MeshServiceSpecEndpoints `json:"endpoints,omitempty"`
}
