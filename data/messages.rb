# This file holds all methods related to accessing the louds message data

require File.dirname(__FILE__) + '/message'
require File.dirname(__FILE__) + "/messages/yaml.rb"

module Louds
module Data

# Factory class for the message container - filename determines which class to actually instantiate:
# * `*.yml` is Louds::Data::Messages::YAML
# * `*.db` is Louds::Data::Messages::SQLite
#
# TODO: Auto-pull all files in messages/ and let them register what filetype they use so one could
# easily "plug in" a new format.
class Messages
  attr_reader :last

  def initialize(filename)
    @file = filename
    @dirty = false
    @messages = {}
    @random_messages = []
    @last = nil
  end

  # Populates the message structure so that random items can be produced
  def load
    @messages.clear
    @random_messages.clear

    raw_messages = retrieve_messages
    if raw_messages.empty?
      raw_messages = [{:text => "ROCK ON WITH SUPERLOUD", :author => "SUPERLOUD"}]
    end

    # Convert from data hash to Message object
    raw_messages.each do |data|
      @messages[data[:text]] = Message.from_hash(data)
      @messages[data[:text]].container = self
    end

    @random_messages.push(*@messages.keys.shuffle)
    @dirty = false
  end

  # Must be implemented in the subclass
  def retrieve_messages
    raise NotImplementedError
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(data)
    string = data[:text]
    return if @messages[string]

    message = Message.new(data)
    @messages[string] = message
    message.container = self

    dirty!
  end

  # Pulls a random message, reloading the data if necessary
  def random
    @random_messages = @messages.keys.shuffle if @random_messages.empty?
    @last = @messages[@random_messages.pop]

    return @last
  end

  def dirty?
    return @dirty
  end

  def dirty!
    @dirty = true
  end

  # Must be implemented in the subclass
  def serialize
    raise NotImplementedError
  end
end

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
