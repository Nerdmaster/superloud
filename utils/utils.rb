# Anything the whole "stack" should include goes here
ROOT = File.expand_path(File.dirname(__FILE__) + "/..")

def lib(*libraries)
  for library in libraries
    require([ROOT, library].join("/"))
  end
end
