package main

// How updates should work:
// - starting the agent for the first time will make it generate a new shared key, stores locally
// and inform it to the server when registering.
// - run "zmon register <email@>" to assign it to a user.
// -- alternative (more secure): "zmon register" and then follow the link it prints.
//
// - Find a better de-dupe method than simply hostname.
