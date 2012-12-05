require "yaml"
require File.expand_path(File.dirname(__FILE__) + '/../messages.rb')

module Louds
module Data
class Messages

class Louds::Data::Messages::YAML < Louds::Data::Messages
  # Returns true if the given message is in our hash
  def exists?(text)
    return true if @messages[text]
    return false
  end

  # YAML-specific method for pulling data - returns an empty array if the YAML file isn't there
  def retrieve_messages
    return FileTest.exist?(@file) ? ::YAML.load_file(@file) : []
  end

  # Stores messages into a YAML file
  def write_data
    # Convert from Message objects to raw hashes
    hashes = []
    @messages.each {|k, v| hashes.push v.to_hash}
    File.open(@file, "w") {|f| f.puts hashes.to_yaml}
  end
end

end
end
end
