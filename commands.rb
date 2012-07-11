# THIS IS WHERE ALL OF MY COMMANDS LIVE PLEASE DO NOT CHANGE STUFF HERE UNLESS IT IS REALLY REALLY
# A DIRECT PART OF COMMANDS OKAY

# Commands we support
VALID_COMMANDS = [
  :dongme, :redongme, :upvote, :downvote, :score, :help, :rps, :biggestdong
]

# RPS stuff is complicated enough to centralize all functionality in here
require "rps/rps_command"

#####
# Command handlers
#####

def biggestdong(e, params)
  if @size_data.empty?
    @irc.msg(e.channel || e.nick, "ONOES NO DONGS TODAY SIRS")
    return
  end

  big_winner = Hash.new(0)
  for user_hash, data in @size_data
    big_winner = data if big_winner[:size] < data[:size]
  end

  nick = big_winner[:nick]
  cm = big_winner[:size]
  inches = cm / 2.54
  @irc.msg(e.channel || e.nick, "THE BIGGEST I'VE SEEN TODAY IS #{nick.upcase}'S WHICH WAS %0.1f INCHES (%d CM)" % [inches, cm])
end

def dongme(e, params)
  send_dong(e)
end

# Increments rerolls and sends inappropriate imagery
def redongme(e, params)
  # Clear old size data for this user
  @size_data[user_hash(e.msg)] = Hash.new(0)

  # Increment mulligan count
  @redongs[user_hash(e.msg)] += 1

  # Send a BRAND NEW DONG!!!
  send_dong(e)
end

# Votes the current message +1
def upvote(e, params)
  vote(1)
end

def downvote(e, params)
  vote(-1)
end

# Reports the last message's score
def score(e, params)
  if !@last_message
    @irc.msg(e.channel || e.nick, "NO LAST MESSAGE OR IT WAS DELETED BY !DOWNVOTE")
    return
  end

  @irc.msg(e.channel || e.nick, "#{@last_message}: #{@messages[@last_message]}")
end

def help(e, params)
  commands = VALID_COMMANDS.collect {|cmd| "!" + cmd.to_s.upcase}.join(" ")
  @irc.msg(e.channel || e.nick, "I HAVE COMMANDS AND THEY ARE: #{commands}")
end

#####
# Command helpers
#####

# Takes an event message and returns a semi-unique hash number representing the user + host
def user_hash(message)
  return message.user.hash + message.host.hash
end

# Computes and seeds RNG with the given event's user hash combined with today's date and optional
# added value, yields to the passed-in block, and then resets the seed to the old value.
def user_seed(user_hash, add)
  old_seed = srand(user_hash + Date.today.strftime("%Y%m%d").to_i + add)
  yield
  srand(old_seed)
end

# Computes size for a given event message.  Stores data into list of sizes for the day.
def compute_size(e)
  user_hash = user_hash(e.msg)
  mulligans = @redongs[user_hash]

  # -2 to size for each !REDONGME command
  size_modifier = mulligans * -2

  # Get size by using daily seed
  size = 0
  user_seed(user_hash, mulligans * 53) do
    size = [2, rand(20).to_i + 8 + size_modifier].max
  end

  @size_data[user_hash] = {:size => size, :nick => e.nick}

  return size
end

# Just keepin' the plagiarism alive, man.  At least in my version, size is always based on requester.
def send_dong(e)
  size = compute_size(e)
  @irc.msg(e.channel || e.nick, "8%sD" % ['=' * size])
end

# Adds +value+ to the score of the last message, if there was one.  If the score goes too low, we
# remove that message forever.
def vote(value)
  return unless @last_message

  @messages[@last_message] += value
  if @messages[@last_message] <= -1
    @messages.delete(@last_message)
    @last_message = nil
  end
  @dirty_messages = true
end
