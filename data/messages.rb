# This file holds all methods related to accessing the louds message data

module Louds
module Data

class Messages
  def initialize
    @messages = {}
    @random_messages = []
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
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(string)
    @messages[string] ||= 1
  end

  # Pulls a random message, reloading the data if necessary
  def random
    @random_messages = @messages.keys.shuffle if @random_messages.empty?
    @last_message = @random_messages.pop
  end
end

end
end
