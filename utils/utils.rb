# Anything the whole "stack" should include goes here
ROOT = File.expand_path(File.dirname(__FILE__) + "/..")

# lib() will auto-pull from root/louds if file isn't found - this is kind of a hack, but louds/ is
# where all internal louds-only files live, so it is a sensible hack (and preserves compatibility
# with prior directory structure)
LOUDS_ROOT = [ROOT, "louds"].join("/")

def lib(*libraries)
  for library in libraries
    begin
      require [ROOT, library].join("/")
    rescue LoadError
      require [LOUDS_ROOT, library].join("/")
    end
  end
end
