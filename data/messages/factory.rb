module Louds
module Data

require File.dirname(__FILE__) + "/../messages.rb"
require File.dirname(__FILE__) + "/yaml.rb"
require File.dirname(__FILE__) + "/sqlite.rb"

class Messages::Factory
  # Generates the appropriate messages object based on file type
  def self.create(filename)
    case File.extname(filename)
      when ".yml", ".yaml"    then return Louds::Data::Messages::YAML.new(filename)
      when ".db", ".sqlite"   then return Louds::Data::Messages::SQLite.new(filename)
    end
  end
end

end
end
