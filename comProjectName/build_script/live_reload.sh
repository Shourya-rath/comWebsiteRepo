templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."

# Backend will be sending files , 
# This means updating the backend will cause updates in the frontend ? 
# will updating non templ files cause updates ? Yes
# seems like a good thing
# TODO: 
# will updating comments also live reload ?
# will updating non Go files also do it ? 
