module Louds
module Data

lib 'data/messages', 'data/messages/yaml', 'data/messages/sqlite'

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
