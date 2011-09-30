# Stores a LOUD message into the hash and responds.
def it_was_loud(message, channel)
  @irc.log.debug "IT WAS LOUD!  #{message.inspect}"

  @messages[message] ||= 1
  random_message(channel)
end

# This is our main message handler.
#
# We store and respond if messages meet the following criteria:
# * It is long (11 characters or more)
# * It has at least one space
# * It has no lowercase letters
# * At least 60% of the characters are uppercase letters
def incoming_message(e)
  # We don't respond to "private" messages
  return if e.pm?

  text = e.message

  return if text =~ /[a-z]/

  # Count various exciting things
  len = text.length
  uppercase_count = text.scan(/[A-Z]/).length
  space_count = text.scan(/\s/).length

  if len >= 11 && uppercase_count >= (len * 0.60) && space_count >= 1
    it_was_loud(e.message, e.channel)
  end
end

# Pulls a random message from our messages array and sends it to the given channel.  Reshuffles
# the main array if the randomized array is empty.
def random_message(channel)
  @random_messages = @messages.keys.shuffle if @random_messages.empty?
  @last_message = @random_messages.pop
  @irc.msg(channel, @last_message)
end

# Just keepin' the plagiarism alive, man.  At least in my version, size is always based on requester.
def send_dong(channel, user_hash)
  mulligans = @redongs[user_hash]
  old_seed = srand(user_hash + (Time.now.to_i / 86400) + mulligans * 53)

  # -2 to size for each !REDONGME command
  size_modifier = mulligans * -2
  size = [2, rand(20).to_i + 8 + size_modifier].max

  @irc.msg(channel, "8%sD" % ['=' * size])
  srand(old_seed)
end

# Adds +value+ to the score of the last message, if there was one.  If the score goes too low, we
# remove that message forever.
def vote(value)
  return unless @last_message

  @messages[@last_message] += value
  if @messages[@last_message] <= -1
    @last_message = nil
    @messages.delete(@last_message)
  end
  @dirty_messages = true
end

# Reports the last message's score
def score(channel)
  if !@last_message
    @irc.msg(channel, "NO LAST MESSAGE OR IT WAS DELETED BY !DOWNVOTE")
    return
  end

  @irc.msg(channel, "#{@last_message}: #{@messages[@last_message]}")
end

# Takes an event message and returns a semi-unique hash number representing the user + host
def user_hash(message)
  return message.user.hash + message.host.hash
end

# Handles a command (string begins with ! - to keep with the pattern, I'm making our loudbot only
# respond to loud commands)
def do_command(command, e)
  case command
    when "REDONGME"
      @redongs[user_hash(e.msg)] += 1
      send_dong(e.channel, user_hash(e.msg))

    when "DONGME"         then send_dong(e.channel, user_hash(e.msg))
    when "UPVOTE"         then vote(1)
    when "DOWNVOTE"       then vote(-1)
    when "SCORE"          then score(e.channel)
    when "HELP"           then @irc.msg(e.channel, "I HAVE COMMANDS AND THEY ARE !DONGME !REDONGME !UPVOTE !DOWNVOTE !SCORE AND !HELP")
  end

  # Here we're saying that we don't want any other handling run - no filters, no handler.  For
  # commands, I put this here because I know I don't want any other handlers having to deal with
  # strings beginning with a bang.
  e.handled!
end
