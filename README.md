### What do you need to run the aggregator?
PostgreSQL, Golang, Linux (but may work on Windows with little prodding)

### Installation
This is Golang so:
[go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)

### Config
The configuration file will be automatically created on first run.

It is located in the user's home directory:
.gatorconfig.json

and its structure is:
```
{
"db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
"current_user_name": "username"
}
```

You may need to manually modify the db_url.
The current_user_name should be empty before one is created.

### Commands?
There are indeed commands. But there is no real help command, and neither will you find an
explanation here. Go (heh) look at the code I guess.

### What is there left to do?
Well...
1. There are several error returns which are natural and expected to happen that should just
   print info on what went wrong instead of dumping the error as is.
2. A comprehensive help command would be nice and an explanation of what command does what.
3. The browse command could use a glow up as well. The way it works now is not terribly useful.
4. There are some warnings that need cleaning up, as well as a bit of general uncleanliness.
5. This README could use a serious rewrite as well.

### Why not do that right now?
The working reasoning is:

I will hopefully come back to it in some time after I forget everything about this project,
so that I may get some more experience dealing with code as such.

The one that is probably most correct:

I don't want to, and I don't have to.
This is my project after all, and I have more pressing things to learn right now.