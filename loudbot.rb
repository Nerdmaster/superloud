require "./data/messages/factory"

# Sets up one-time data when loudbot starts
def init_data
  # Initialize message data object
  @messages = Louds::Data::Messages::Factory.create("loud_messages.yml")
  @messages.load

  @channel_list = []

  # RPS competitional fun
  @rps_data ||= {}
  @rps_contestant = {}

  # Ignore list
  @ignore_regexes = []
  ignores = FileTest.exist?("ignores.txt") ? IO.readlines("ignores.txt") : []
  for ignore in ignores
    @ignore_regexes.push Regexp.new(ignore.strip, Regexp::IGNORECASE)
  end
end

# Sets up data on init and each day
def init_daily_data
  @size_data = {}
  @redongs = Hash.new(0)
  @last_ping_day = Date.today
end

# Given the text, checks various conditions, and calls the block with a status to be used in
# determining if the message should be ignored, responded to, and/or saved
def is_it_loud?(text, &block)
  errors = []

  # No lowercase
  errors.push("includes lowercase letters") if text =~ /[a-z]/

  # Count various exciting things
  len = text.length
  words = text.split(/[^A-Z']+/)
  letters = text.scan(/[A-Z]/)
  uppercase_count = letters.length

  # Rules are getting complex - let's handle them one at a time
  errors.push("too short") if len < 11
  errors.push("too low uppercase ratio") if uppercase_count < len * 0.60
  errors.push("shut up") if (text =~ /retard/ || text =~ /reetard/)

  # If there are any errors, the text is bad and nothing else needs to happen
  unless errors.empty?
    yield(:bad, errors.join(", "))
    return
  end

  # Throw out any words that are obvious nonsense
  nonsense = []
  for word in words.dup
    # I and A are valid words
    next if word.length == 1

    # No vowels?  NO WORD!
    nonsense.push(words.delete(word)) if word =~ /^[^AEIOUY]+$/

    # Vowels only?  NO WORD!!
    nonsense.push(words.delete(word)) if word =~ /^[AEIOU]+$/
  end

  # Fewer than 2 unique words *or* half the words aren't unique?  GTFO!
  errors.push("too few unique words") if words.uniq.count < 2 || words.uniq.count <= words.count / 2

  # What about letters per word?!?
  errors.push("words are too small") if (words.count > 0) && (letters.count / words.count < 3.0)

  unless errors.empty?
    yield(:rejected, errors.join(", "))
    return
  end

  # WE MADE IT OMG
  yield(:loud, "loud")
end

# This is our main message handler.
#
# We store and respond if messages meet the following criteria:
# * It is long (11 characters or more)
# * It has at least one space
# * It has no lowercase letters
# * At least 60% of the characters are uppercase letters
# * Words aren't "invalid":
#   * Average letters per word is 3 or more
#   * A .uniq call should not cut more than 1/2 the words
def incoming_message(e)
  # We don't respond to "private" messages
  return if e.pm?

  text = e.message

  # Check all the various conditions for responding to the message and including the message
  is_it_loud?(text) do |status, reason|
    case status
      when :bad
        @irc.log.debug "Ignoring message: #{reason}"
        return

      when :rejected
        random_message(e.channel)
        @irc.log.debug "Rejecting message: #{reason}"

      when :loud
        @irc.log.debug "IT WAS LOUD!  #{text.inspect}"
        random_message(e.channel)
        @messages.add(:text => text, :author => e.nick)
    end
  end

end

# Pulls a random message from our messages array and sends it to the given channel
def random_message(channel)
  message = @messages.random
  @irc.msg(channel, message.text)
  message.view!
end

# Handles a command (string begins with ! - to keep with the pattern, I'm making our loudbot only
# respond to loud commands)
def do_command(e, command, params)
  # Empty command means no command
  return if command.empty?

  # Convert to a method symbol
  command = command.downcase.to_sym

  # If valid, send it off
  if VALID_COMMANDS.include?(command)
    method = self.method(command)
    method.call(e, params)
  end

  # Here we're saying that we don't want any other handling run - no filters, no handler.  For
  # commands, I put this here because I know I don't want any other handlers having to deal with
  # strings beginning with a bang.
  e.handled!
end
