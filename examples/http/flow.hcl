endpoint "FetchLatestProject" "http" {
	endpoint = "/"
	method = "GET"
	codec = "json"
}

flow "FetchLatestProject" {
	input "placeholder.Query" {}

	resource "query" {
		request "placeholder.Service" "GetTodo" {
		}
	}

	resource "user" {
		request "placeholder.Service" "GetUser" {
		}
	}

	output "placeholder.Item" {
		header {
			Username = "{{ user:username }}"
		}

		id = "{{ query:id }}"
		title = "{{ query:title }}"
		completed = "{{ query:completed }}"
	}
}
