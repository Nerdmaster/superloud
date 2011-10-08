# THIS IS WHERE ALL OF MY COMMANDS LIVE PLEASE DO NOT CHANGE STUFF HERE UNLESS IT IS REALLY REALLY
# A DIRECT PART OF COMMANDS OKAY

# Commands we support
VALID_COMMANDS = [
  :dongme, :redongme, :upvote, :downvote, :score, :help
]

#####
# Command handlers
#####

def dongme(e)
  send_dong(e)
end

# Increments rerolls and sends inappropriate imagery
def redongme(e)
  @redongs[user_hash(e.msg)] += 1
  send_dong(e)
end

# Votes the current message +1
def upvote(e)
  vote(1)
end

def downvote(e)
  vote(-1)
end

# Reports the last message's score
def score(e)
  if !@last_message
    @irc.msg(e.channel, "NO LAST MESSAGE OR IT WAS DELETED BY !DOWNVOTE")
    return
  end

  @irc.msg(e.channel, "#{@last_message}: #{@messages[@last_message]}")
end

def help(e)
  commands = VALID_COMMANDS.collect {|cmd| "!" + cmd.to_s.upcase}.join(" ")
  @irc.msg(e.channel, "I HAVE COMMANDS AND THEY ARE: #{commands}")
end

#####
# Command helpers
#####

# Takes an event message and returns a semi-unique hash number representing the user + host
def user_hash(message)
  return message.user.hash + message.host.hash
end

# Just keepin' the plagiarism alive, man.  At least in my version, size is always based on requester.
def send_dong(e)
  channel = e.channel
  user_hash = user_hash(e.msg)
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
