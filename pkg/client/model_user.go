/*
 * Grendel API
 *
 * Bare Metal Provisioning system for HPC Linux clusters. Find out more about Grendel at [https://github.com/ubccr/grendel](https://github.com/ubccr/grendel)
 *
 * API version: 1.0.0
 * Contact: aebruno2@buffalo.edu
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client
// User struct for User
type User struct {
	Id string `json:"id,omitempty"`
	Username string `json:"username"`
	PasswordHash string `json:"password_hash,omitempty"`
	Role string `json:"role,omitempty"`
}
