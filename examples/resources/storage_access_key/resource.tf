resource "airtelcloud_storage_access_key" "application" {
  # Keep the access key valid for 30 days.
  expiry = 2592000
}
