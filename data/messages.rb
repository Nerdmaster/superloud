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
end

end
end
