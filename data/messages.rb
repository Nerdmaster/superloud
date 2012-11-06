# This file holds all methods related to accessing the louds message data

require "singleton"

module Louds
module Data

class Message
  attr_reader :author, :views, :score, :text

  def initialize(text, author)
    @@messages ||= Louds::Data::Messages.instance

    @text = text
    @author = author
    @views = 0
    @score = 1
  end

  def view!
    @views += 1
    @@messages.dirty!
  end

  def upvote!
    @score += 1
    @@messages.dirty!
  end

  def downvote!
    @score -= 1
    @@messages.dirty!
  end
end

class Messages
  include Singleton

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
    @messages = FileTest.exist?(@@file) ? YAML.load_file(@@file) :
                {"ROCK ON WITH SUPERLOUD" => Message.new("ROCK ON WITH SUPERLOUD", "SUPERLOUD")}
    @random_messages = @messages.keys.shuffle
    @dirty = false
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(string, author)
    @messages[string] ||= Message.new(string, author)
    dirty!
  end

  # Pulls a random message, reloading the data if necessary
  def random
    @random_messages = @messages.keys.shuffle if @random_messages.empty?
    @last = @messages[@random_messages.pop]
    @last.view!

    return @last.text
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

    File.open(@@file, "w") {|f| f.puts @messages.to_yaml}
    @dirty = false
  end
end

end
end
