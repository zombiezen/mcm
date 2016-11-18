using Cxx = import "/third_party/capnproto/c++/src/capnp/c++.capnp";
using Go = import "/third_party/golang/capnproto/std/go.capnp";

@0x88881ae7af33bcb5;
$Cxx.namespace("mcm");
$Go.package("catalog");
$Go.import("github.com/zombiezen/mcm/catalog");

struct Catalog {
  # The root struct in a catalog file.

  resources @0 :List(Resource);
}

using ResourceId = UInt64;

struct Resource {
  id @0 :ResourceId $Go.name("ID");
  # The resource's identifier, used for dependencies.
  # The identifier should be unique within a catalog.

  comment @1 :Text;
  # An optional human-readable description of the resource for use in
  # error and progress messages.

  dependencies @2 :List(ResourceId);
  # Resources that must be applied before this resource can be applied.

  union {
    noop @3 :Void;
    # Does nothing.  Mainly to give the resource a safe default.

    file @4 :File;
    exec @5 :Exec;
  }
}

struct File @0x8dc4ac52b2962163 {
  # An entry on the filesystem.

  path @0 :Text;
  # An absolute OS-specific file path.

  union {
    plain :group {
      content @1 :Data;
      # Byte content of the file.  If null, then file content is
      # untouched by the executor, but it is an error if the file does
      # not exist.

      mode @2 :Mode;
    }
    directory :group {
      mode @3 :Mode;
    }

    absent @4 :Void;
  }

  struct Mode {
    user @0 :UserRef;
    group @1 :GroupRef;
  }
}

struct UserRef {
  # A reference to a OS user.

  union {
    id @0 :Int32;
    name @1 :Text;
  }
}

struct GroupRef {
  # A reference to an OS group.

  union {
    id @0 :Int32;
    name @1 :Text;
  }
}

struct Exec @0x984c97311006f1ca {
  # An execution of an arbitrary program.

  struct Command {
    # An individual program invocation.

    union {
      argv @0 :List(Text);
      # A list of arguments as passed to exec.
      # There must be at least one argument, which must be an absolute
      # path to the executable.

      bash @1 :Text;
      # A script as passed to bash.
    }

    struct EnvVar {
      # A single environment variable.

      name @0 :Text;
      value @1 :Text;
    }

    environment @2 :List(EnvVar);
    # The subprocess's environment.
    # An empty or null list is an empty environment.

    workingDirectory @3 :Text;
    # The subprocess's working directory.
    # An empty or null string is the root.
  }

  command @0 :Command;
  # Command to run if the condition is met.

  condition :union {
    always @1 :Void;
    # Command will always be run.  Used if the command is idempotent.
    onlyIf @2 :Command;
    # Command will be run only if another command returns a successful exit code.
    # It is assumed that this command does not do anything destructive and may be run multiple times.
    unless @3 :Command;
    # Command will be run only if another command returns a failure exit code.
    # It is assumed that this command does not do anything destructive and may be run multiple times.
    fileAbsent @4 :Text;
    # Command will be run only if the OS file path does not exist.
  }
}
