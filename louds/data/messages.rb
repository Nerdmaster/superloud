# This file holds all methods related to accessing the louds message data

lib 'data/message'

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
    @new_messages = []
    @changed_messages = []
    @last = nil
    @voted = {}
  end

  # Populates the message structure so that random items can be produced
  def load
    @messages.clear
    @random_messages.clear
    @new_messages.clear
    @changed_messages.clear

    raw_messages = retrieve_messages
    if raw_messages.empty?
      raw_messages = [{:text => "ROCK ON WITH SUPERLOUD", :author => "SUPERLOUD"}]
    end

    # Convert from data hash to Message object
    raw_messages.each {|data| new_message!(data)}

    @random_messages.push(*@messages.keys.shuffle)
    @dirty = false
  end

  # Must be implemented in the subclass
  def retrieve_messages
    raise NotImplementedError
  end

  # Returns true if the given message text already exists - must be overridden at the subclass
  def exists?(text)
    raise NotImplementedError
  end

  # Stores the given string if it isn't already stored, setting the score to 1
  def add(data)
    return if exists?(data[:text])
    dirty!(:new => new_message!(data))
  end

  # Creates a new message object using the given data hash and stores it in our container hash.
  def new_message!(data)
    message = Message.new(data)
    message.container = self
    @messages[message.text] = message
    return message
  end

  # Pulls a random message, reloading the data if necessary
  def random
    @random_messages = @messages.keys.shuffle if @random_messages.empty?
    @last = @messages[@random_messages.pop]
    @voted.clear

    return @last
  end

  # Casts a vote for the last message if the given event's user hasn't already voted for it this
  # time around.  Small databases can get crazy voting, but larger databases will be fine.
  #
  # TODO: If this is adjusted to disallow users for reasons other than having already voted, we
  # need to pass something back to the caller because right now the message is very specific to
  # duplicate voting.
  def vote(user_hash, value)
    return false if @voted[user_hash]

    case value
      when 1  then @last.upvote!
      when -1 then @last.downvote!
    end

    @voted[user_hash] = true
    return true
  end

  def dirty?
    return @dirty
  end

  def dirty!(changes = {})
    @dirty = true

    for type, items in changes
      case type
        when :new     then @new_messages.push(*items)
        when :changed then @changed_messages.push(*items)
      end
    end

    @changed_messages.uniq!
  end

  # Calls write_data and clears all dirty state info
  def serialize
    return unless dirty?

    write_data

    @dirty = false
    @new_messages.clear
    @changed_messages.clear
  end

  # Must be implemented in the subclass
  def write_data
    raise NotImplementedError
  end
end

end
end
