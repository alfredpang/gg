gg - golang go get vendor manager
==================================

If you use golang, you will have to deal with dependency and vendoring at some point. The authors of golang describes their recommended approach in [this google group discussion post](https://groups.google.com/forum/#!topic/golang-dev/nMWoEAG55v8%5B1-25%5D). gg aims to be a tool to help automate the work described in the post.

To summarize the vendoring approach:

*  Designate a package prefix (and directory) to be the place to save packages that you would like to vendor.
*  When vendoring a package, be sure to vendor all the necessary dependencies.
*  Track revisions of said packages.
*  When using the vendored packages, reference the vendored location.

That said, this tool allows for per project vendoring (say into an "internal" directory), or a shared vendoring approach.

Usage
-----

```
gg [command] [options] [packages]

Manage vendor directory:

 vinit    Initialize vendor directory.
 vadd     Add package to vendor.
 vlist    List packages being vendored.
 vrebuild Rebuild from config file.
 vupdate  Update packages.

Import rewriting:

 usev     Rewrite to use import vendored packages.
 unusev   Rewrite to undo vendored imports.

Other commands:

 rdep     List dependencies of a go-getable package.
 ldep     List dependencies of local directory/package.
 listcore List known core packages.

Special files:

 _ggv.json Vendor configuration file.
 .gg       Specifies vendor root to use.

Use "gg help <command>" for usage of a specific command.
```

Vendor directory uses _ggv.json to track vendored packages. Your project that uses the vendoring may specify a .gg file to specify the vendoring root to use.

How to Use
----------
If you are starting a new project, you should work as usual using go get. Once you are satisfied with your packages, run "gg ldep" in order to figure out the dependencies of your project. From there create a vendoring directory and use vadd to add the necessary packages. Then create a ".gg" file and use "gg usev" to rewrite your canonical imports to your vendored imports.

If you have an existing project, run "gg ldep" and then follow the same steps as a new project.

You may use "gg rdep" to evaluate the dependencies of a package without modifying your local GOPATH workspace.

You should keep your vendored directory and your _ggv.json under source control. That way, you may run "gg vupdate" and rebuild the world to test the updates, before moving your vendored packages forward.

Cavets
------
* Supports one GOPATH path only.

Examples
--------
Assuming GOPATH has already been set to ~/go_work

Simple vendoring in "internal" directory. This will vendor github.com/gorilla/mux as well as dependency github.com/gorilla/context. After usev, myproj's imports will be rewritten to use myproj/github.com/gorilla/mux.
```
> cd go_work/src/myproj
> mkdir internal
> cd ..
> gg vinit
> gg vadd github.com/gorilla/mux
> cd go_work/src/myproj
> gg usev
```

Your vendoring directory does not have to live in a subdirectory.
```
> cd go_work/src
> mkdir v
> cd v
> gg vinit
> gg vadd package1 package2 package3
> cd go_work/src/myproj
> echo v > .gg
> gg usev
```

Perhaps your package does not implement the go get protocol. You may specify the repo details directly.
```
> cd go_work/src/v
> gg vadd -vcs hg -vcs-source https://code.google.com/p/go-charset -notes "can't go get this but we can pull it with hg" code.google.com/p/go-charset
```


