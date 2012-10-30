# This file holds all methods related to accessing the louds message data

module Louds
module Data

class Messages
  attr_reader :last

  def initialize
    @dirty = false
    @messages = {}
    @random_messages = []
    @last = nil
  end

  # Populates the message structure so that random items can be produced
  def load
    # Loud messages can be newline-separated strings in louds.txt or an array or hash serialized in
    # louds.yml.  If messages are an array, we convert all of them to hash keys with a score of 1.
    @messages = FileTest.exist?("louds.yml") ? YAML.load_file("louds.yml") :
                FileTest.exist?("louds.txt") ? IO.readlines("louds.txt") :
                {"ROCK ON WITH SUPERLOUD" => 1}
    if Array === @messages
      dupes = @messages.dup
      dupes.each {|string| @messages[string.strip] = 1}
    end

    @random_messages = @messages.keys.shuffle
    @dirty = false
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(string)
    @messages[string] ||= 1
    @dirty = true
  end

  # Pulls a random message, reloading the data if necessary
  def random
    @random_messages = @messages.keys.shuffle if @random_messages.empty?
    @last = @random_messages.pop
  end

  def dirty?
    return @dirty
  end

  # Stores messages into a YAML file
  def serialize
    return unless dirty?

    File.open("louds.yml", "w") {|f| f.puts @messages.to_yaml}
    @dirty = false
  end

  # Adds +value+ to the score of the last message, if there was one.  If the score goes too low, we
  # remove that message forever.
  def vote(value)
    return unless @last

    @messages[@last] += value
    if @messages[@last] <= -1
      @messages.delete(@last)
      @last = nil
    end
    @dirty = true
  end

  # Returns the score of the last message
  def last_score
    return @messages[@last]
  end
end

end
end
