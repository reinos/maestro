endpoint "user" "http" {
	endpoint = "/v1/user"
	method = "GET"
	codec = "json"
}

flow "user" {
	resource "query" {
		request "proto.Users" "GetUser" {
		}
	}

	output "proto.User" {
        id = "{{ query:id }}"
        name = "{{ query:name }}"
        username = "{{ query:username }}"
    }
}