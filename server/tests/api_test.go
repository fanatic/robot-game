package tests

import (
	"testing"
)

func TestAPI(t *testing.T) {
	setup(t)
	defer teardown()

	t.Run("basic client", func(t *testing.T) {
		assertResponse(t, GET(t, "/state"),
			`{"grid": 16, "robots": [], "round": 0}`, 200)

		id := assertResponse(t, POST(t, "/robots", `{"name": "JP"}`),
			`{
				"dead":false, 
				"x":1, 
				"y":15, 
				"score":0, 
				"name":"JP", 
				"color":"#e6194b", 
				"direction":3, 
				"vision":4
			}`, 200)

		assertResponse(t, GET(t, "/state"),
			`{
				"grid": 16, 
				"robots": [{
					"dead":false, 
					"x":1, 
					"y":15, 
					"score":0, 
					"name":"JP", 
					"color":"#e6194b", 
					"direction":3, 
					"vision":4
				}], 
				"round": 0
			}`, 200)

		assertResponse(t, GET(t, "/robots/"+id),
			`{
				"dead":false, 
				"x":1, 
				"y":15, 
				"score":0, 
				"name":"JP", 
				"color":"#e6194b", 
				"direction":3, 
				"vision":4
			}`, 200)

		assertResponse(t, POST(t, "/robots/"+id+"/move", ``),
			`{
				"dead":false, 
				"x":0, 
				"y":15, 
				"score":0, 
				"name":"JP", 
				"color":"#e6194b", 
				"direction":3, 
				"vision":4
			}`, 200)

		assertResponse(t, POST(t, "/robots/"+id+"/turn", `{"direction": false}`),
			`{
				"dead":false, 
				"x":0, 
				"y":15, 
				"score":0, 
				"name":"JP", 
				"color":"#e6194b", 
				"direction":0, 
				"vision":4
			}`, 200)

		assertResponse(t, POST(t, "/robots/"+id+"/move", ``),
			`{
				"dead":false, 
				"x":0, 
				"y":14, 
				"score":0, 
				"name":"JP", 
				"color":"#e6194b", 
				"direction":0, 
				"vision":4
			}`, 200)

		assertResponse(t, POST(t, "/robots/"+id+"/attack", ``),
			`{
				"at":"error", 
				"msg": "swwwing and a missss"
			}`, 200)
	})
}
