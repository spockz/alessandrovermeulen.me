--- 
layout: post
title: Debian Lenny and the Haskell Platform
tags: 
- Haskell
- Haskell platform
- debian
- lenny
comments: true
---
Currently there is no package for the Haskell-Platform in Debian stable. However, the source of this platform and GHC is available for download at <a href="http://www.haskell.org/ghc/">ghc</a> and <a href="http://www.haskell.org/platform">platform</a>.

However, there are some issues to solve when installing the platform from source. Mainly you'll be missing several packages. The following commands worked at my own Debian machine. If you find out that's something is missing that was apparently already installed on my machine, don't hesitate to leave a comment.

The following commands are all executed as <em>root</em>.
First we need to download the sources, I'm using the x86 sources, but feel free to use the x64 version:
[bash]
cd /tmp
wget http://www.haskell.org/ghc/dist/6.10.4/ghc-6.10.4-i386-unknown-linux-n.tar.bz2
wget http://hackage.haskell.org/platform/2009.2.0.2/haskell-platform-2009.2.0.2.tar.gz
[/bash]

Now we need to extract these sources:

[bash]
tar xvf ghc-6.10.4-i386-unknown-linux-n.tar.bz2
tar xvzf haskell-platform-2009.2.0.2.tar.gz
[/bash]

Before we do any other stuff, let's ensure that we have all the packages and libraries we need before continuing.

[bash]
apt-get install build-essential libghc6-opengl-dev libghc6-glut-dev libeditline-dev libedit2 libedit-dev
[/bash]

Also make sure that you apply this patch to the platform, the bug has been around for some time now but hasn't been fixed apparently, it can be found at <a href="http://trac.haskell.org/haskell-platform/ticket/84">bug #84</a>:

[bash]
patch -p0 haskell-platform-2009.2.0.2/scripts/install.sh
Index: haskell-platform-2009.2.0.2/scripts/install.sh
===================================================================
--- haskell-platform-2009.2.0.2.orig/scripts/install.sh
+++ haskell-platform-2009.2.0.2/scripts/install.sh
@@ -34,13 +34,23 @@ install_pkg () {
   fi
 }
 
+# Is this exact version of the package already installed?
+is_pkg_installed () {
+  PKG_VER=$1
+  grep &quot; ${PKG_VER} &quot; installed.packages &gt; /dev/null 2&gt;&amp;1
+}
+
 # Actually do something!
 cd packages
 for pkg in `cat platform.packages`; do
-  cd &quot;${pkg}&quot; || die &quot;The directory for the component ${PKG} is missing&quot;
-  echo &quot;Installing ${pkg}...&quot;
-  install_pkg ${pkg}
-  cd ..
+  if is_pkg_installed &quot;${pkg}&quot;; then
+    echo &quot;Platform package ${pkg} is already installed. Skipping...&quot;
+  else
+    cd &quot;${pkg}&quot; || die &quot;The directory for the component ${PKG} is missing&quot;
+    echo &quot;Installing ${pkg}...&quot;
+    install_pkg ${pkg}
+    cd ..
+  fi
 done
 
 echo
[/bash]

Now go to the ghc dir and build it.

[bash]
cd ghc-6.10.4
./configure
make install
[/bash]

And if all of this goes well you can now go to the haskell-platform dir and install it:

[bash]
cd ../haskell-platform-2009.2.0.2/
./configure
make
make install
[/bash]

And done! :)
