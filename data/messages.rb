# This file holds all methods related to accessing the louds message data

require "singleton"
require "yaml"
require File.dirname(__FILE__) + '/message'

module Louds
module Data

class Messages
  attr_reader :last

  @@file = "loud_messages.yml"

  def initialize
    @dirty = false
    @messages = {}
    @random_messages = []
    @last = nil
  end

  # Populates the message structure so that random items can be produced
  def load
    # Louds messages are now a complex data structure that contains text, author, score, and
    # times viewed.  There is no conversion, sorry.  The messages are still stored in a hash with
    # the text as the index as this allows easier lookups (until we go fully db).
    raw_messages = FileTest.exist?(@@file) ? YAML.load_file(@@file) :
                {"ROCK ON WITH SUPERLOUD" => Message.new("ROCK ON WITH SUPERLOUD", "SUPERLOUD")}

    # Convert from data hash to Message object
    raw_messages.each do |data|
      @messages[data[:text]] = Message.from_hash(data)
      @messages[data[:text]].container = self
    end

    @random_messages = @messages.keys.shuffle
    @dirty = false
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(string, author)
    @messages[string] ||= Message.new(string, author)
    @messages[string].container = self
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

  # Stores messages into a YAML file
  def serialize
    return unless dirty?

    # Convert from Message objects to raw hashes
    hashes = []
    @messages.each {|k, v| hashes.push v.to_hash}
    File.open(@@file, "w") {|f| f.puts hashes.to_yaml}
    @dirty = false
  end
end

end
end
