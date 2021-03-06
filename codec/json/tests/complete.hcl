flow "complete" {
	input "complete" {}

	resource "first" {
		request "mock" "complete" {
			message = "{{ input:message }}"

			message "nested" {
				value = "{{ input:nested.value }}"
			}

			repeated "repeating" "input:repeating" {
				value = "{{input:repeating.value}}"
			}
		}
	}
}