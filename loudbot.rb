require "./data/messages"

# Sets up one-time data when loudbot starts
def init_data
  # Initialize message data object
  @messages = Louds::Data::Messages.new
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
  @last_message = nil
  @dirty_messages = false
  @redongs = Hash.new(0)
  @last_ping_day = Date.today
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

  return if text =~ /[a-z]/

  # Count various exciting things
  len = text.length
  words = text.split(/[^A-Z']+/)
  letters = text.scan(/[A-Z]/)
  uppercase_count = letters.length

  # Rules are getting complex - let's handle them one at a time
  return if len < 11
  return if uppercase_count < len * 0.60
  return if text =~ /retard/
  return if text =~ /reetard/

  # At this point it's close enough, so let's at least spit out a response
  random_message(e.channel)

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

  @irc.log.debug "Rejected #{nonsense.inspect} as nonsense words" unless nonsense.empty?

  # Always exit if we only ended up with one unique word
  if words.uniq.count == 1
    @irc.log.debug "Rejected #{words.inspect} - only one unique word"
    return
  end

  # Half the words aren't unique?  GTFO!
  if words.uniq.count <= words.count / 2
    @irc.log.debug "Rejected #{words.inspect} - too few unique words"
    return
  end

  # What about letters per word?!?
  if letters.count / words.count < 3.0
    @irc.log.debug "Rejected #{text.inspect} - words are too small"
    return
  end

  # WE MADE IT OMG
  @irc.log.debug "IT WAS LOUD!  #{text.inspect}"
  @messages.add(text)
end

# Pulls a random message from our messages array and sends it to the given channel
def random_message(channel)
  @irc.msg(channel, @messages.random)
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
