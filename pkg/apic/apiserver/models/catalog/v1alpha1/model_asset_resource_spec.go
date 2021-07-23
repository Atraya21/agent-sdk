/*
 * API Server specification.
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: SNAPSHOT
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1alpha1

// AssetResourceSpec struct for AssetResourceSpec
type AssetResourceSpec struct {
	// The Stage this Asset Resource is deployed on.
	Stage                  string `json:"stage,omitempty"`
	AssetRequestDefinition string `json:"assetRequestDefinition,omitempty"`
	Type                   string `json:"type"`
	// content-type of the spec.
	ContentType string `json:"contentType,omitempty"`
	// Base64 encoded value of the api specification.
	Definition string `json:"definition"`
	// Resource availabiltiy
	Status string `json:"status"`
	// information to access the definition.
	AccessInfo []AssetResourceSpecAccessInfo `json:"accessInfo,omitempty"`
}
