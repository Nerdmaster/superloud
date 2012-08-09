# Sets up one-time data when loudbot starts
def init_data
  # Loud messages can be newline-separated strings in louds.txt or an array or hash serialized in
  # louds.yml.  If messages are an array, we convert all of them to hash keys with a score of 1.
  @messages = FileTest.exist?("louds.yml") ? YAML.load_file("louds.yml") :
              FileTest.exist?("louds.txt") ? IO.readlines("louds.txt") :
              {"ROCK ON WITH SUPERLOUD" => 1}
  if Array === @messages
    dupes = @messages.dup
    @messages = {}
    dupes.each {|string| @messages[string.strip] = 1}
  end

  @random_messages = @messages.keys.shuffle

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

# Stores a LOUD message into the hash and responds.
def it_was_loud(message, channel)
  @irc.log.debug "IT WAS LOUD!  #{message.inspect}"

  @messages[message] ||= 1
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
  it_was_loud(e.message, e.channel)
end

# Pulls a random message from our messages array and sends it to the given channel.  Reshuffles
# the main array if the randomized array is empty.
def random_message(channel)
  @random_messages = @messages.keys.shuffle if @random_messages.empty?
  @last_message = @random_messages.pop
  @irc.msg(channel, @last_message)
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
