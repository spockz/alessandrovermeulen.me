---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "JCU App installation script",
  "date": "2012-02-29T18:17:00+01:00",
  "slug": "jcu-app-installation-script",
  "draft": false,
  "categories": [
    "haskell"
  ],
  "url": "/2012/02/29/jcu-app-installation-script/",
  "aliases": [
    "/posts/jcu-app-installation-script/"
  ]
}
---

With this script it should be possible to install and build the JCU app in a
local directory. It does *not* build the UHC for you. If that is wanted the
option could be build in of course!

`install.rb`

```ruby
#! /usr/bin/ruby
require 'readline'

def command?(command)
  system("which #{ command} > /dev/null 2>&1")
end

def library?(lib) 
  system("ghc-pkg list | grep #{lib} > /dev/null 2>&1")
end

def request(question, onYes)
  stty_save = `stty -g`.chomp
  trap('INT') { system('stty', stty_save); exit 1 }

  line = Readline.readline(question, true)

  if line.downcase == "y" || line == ""
    return onYes.call
  end

  return false
  
end

def colorize(text, color_code)
  "#{color_code}#{text}\033[0m"
end

def red(text); colorize(text, "\033[31m"); end
def green(text); colorize(text, "\033[32m"); end
def yellow(text); colorize(text, "\033[33m"); end
def blue(text); colorize(text, "\033[34m"); end

class Message
  def self.ok(message)
    puts("[  #{green("CHECK")}  ] #{message}")
  end 

  def self.error(message)
    puts("[  #{red("FAILED")} ] #{message}")
  end
  
  def self.info(message)
    puts("[  #{yellow("INFO")}   ] #{message}")
  end
  
  def self.question(message)
    puts("[  #{blue("REQ")}    ] #{message}")
  end
end

# Message.ok ""
# Message.error ""
# Message.info ""
# Message.question ""
# exit

def check?(condition, message) 
  if condition
    Message.ok(message)
  else
    Message.error(message)
  end
  condition
end

if ENV['UHC'] == ""
  Message.error "Please provide the $UHC variable before running this script!"
  exit 1
end


requiredCommands = ["git", "ghc", "cabal", "uuagc"]
requiredLibs     = ["uulib", "fgl", "network", "binary"]

cabalInstall = "cabal install --user "

requiredCommands.each { | command | 
  if !check?((command? command), command)
    exit 1
  end
}

requiredLibs.each {|lib|
  command = cabalInstall + lib
  if !check?((library? lib), lib)
    if !request(blue("Install #{lib}? [Y/n]: "), lambda { system command })
      exit 1
    end
    !check?((library? lib), lib)
  end
}


# Check Cabal packages

# Check github projects
projects = {"jcu"        => "https://github.com/UU-ComputerScience/JCU",
            "uhcjs"      => "https://github.com/UU-ComputerScience/uhc-js",
            "NanoProlog" => "https://github.com/UU-ComputerScience/NanoProlog"}

# Check whether dir exists and if not pull it
projects.each { |key, path|
  Message.info("Checking " + key)
  if File.directory? key
    command = "cd " + key + " && " + "git pull"
  
  else
    command = "git clone " + path + " " + key
  end
  
  if !system command
    exit 1
  end
}        

if !File.directory? "uu-tc-2009.2.2"
  puts "Downloading uu-tc" 
  `curl -L -O http://www.cs.uu.nl/docs/vakken/b3tc/dist/uu-tc-2009.2.2.tar.gz`
  puts "Extracting it"
  `tar xvzf uu-tc-2009.2.2.tar.gz`
else
  Message.ok "uu-tc library already downloaded."
end
           
# Check snap version
# 

# CD into JCU and try install
Message.info "Installing JCU"
system "cd jcu && cabal install --user"

Message.info "Please make sure that you edit the connection string! And that PostgreSQL is installed!"

if ENV['EDITOR'] != '' 
  Message.info "- Opening the config file"
  system "$EDITOR jcu/config/connection_string.conf"
end

cwd = Dir.pwd

# CD into the Clientside app and build
Message.info "Please provide the $UHC variable yourself: export UHC=..."
Message.info "You can build UHC-JS with the following command:"
Message.info "    make uhc && make 101/ehclib EHC_VARIANT_TARGET=js"
Message.info "Building the client-side "
system "export UHC_JSCRIPT=#{cwd}/uhcjs/uhc-js/src;export UHC_NANOPROLOG=#{cwd}/NanoProlog/src;export UHC_UU_TC=#{cwd}/uu-tc-2009.2.2/src;cd jcu/resources/static/hjs && make"

Message.info "If all went okay you can now cd into the #{yellow "`jcu`"} directory, execute the command jcu and browse to http://localhost:8000."
```
