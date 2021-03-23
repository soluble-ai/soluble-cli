api_prefix = "/foo"
command "group" "foo" {
  short = "Grouping of commands"
  command "print_client" "ping" {
	  short = "ping server"
	  method = "GET"
	  path = "ping/{dummyID}"
	  parameter "dummyID" {
		  usage = "dummy value"
		  disposition = "context"
	  }
	  parameter "action" {
		  usage = "action"
	  }
	  result {
		  path = [ "data" ]
		  columns = [ "col1", "col1" ]
	  }
  }
}