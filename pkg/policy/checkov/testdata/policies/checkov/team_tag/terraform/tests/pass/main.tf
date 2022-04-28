resource "aws_security_group" "main" {
    description = "Test"
    tags = {
        Team = "Team 1"
    }
}