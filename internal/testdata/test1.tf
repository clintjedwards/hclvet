resource "google_compute_instance" "example" {
	provider            = "google"
	name                = "basecoat"
	machine_type        = "f1-micro"
	deletion_protection = "true"
	hostname            = "basecoat.clintjedwards.com"
	metadata            = {}
	labels = {
		"basecoat" = ""
	}
	tags = [
		"basecoat"
	]
}
resource "aws_compute_instance" "not_example" {
}
lolwut = "weow"
happy = "test1"
shouldnt_matter = "lolwut"
