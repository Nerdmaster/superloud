# THIS IS WHERE ALL OF MY COMMANDS LIVE PLEASE DO NOT CHANGE STUFF HERE UNLESS IT IS REALLY REALLY
# A DIRECT PART OF COMMANDS OKAY

require "digest"

# Commands we support
VALID_COMMANDS = [
  :dongwinners, :dwall, :dongrankme, :dongme, :redongme, :upvote, :downvote, :score, :help, :rps, :biggestdong, :size, :sizeme,
  :refresh_ignores, :omakase
]

PLACES = ["FIRST", "SECOND", "THIRD", "FOURTH", "FIFTH", "SIXTH", "SEVENTH", "EIGHTH", "NINTH", "TENTH"]

# RPS stuff is complicated enough to centralize all functionality in here
lib "rps/rps_command"
lib "lib/dice"

#####
# Command handlers
#####

def biggestdong(e, params)
  if @size_data.empty?
    @irc.msg(e.channel || e.nick, "ONOES NO DONGS TODAY SIRS")
    return
  end

  winners = rank_by_size.first
  nametext = userlist_text(winners.collect {|user| user[:nick]}).upcase
  size = winners.first[:size]

  cm = size / 2.0
  inches = cm / 2.54

  if (winners.length <=1)
    @irc.msg(e.channel || e.nick, "THE BIGGEST I'VE SEEN TODAY IS #{nametext}'S WHICH IS %0.1f INCHES (%0.1f CM)" % [inches, cm])
  else
    @irc.msg(e.channel || e.nick, "THE BIGGEST I'VE SEEN TODAY IS... OMFG... IT'S A TIE BETWEEN #{nametext}!  THEY'RE %s %0.1f INCHES (%0.1f CM)!!!1!!" % [(winners.length == 2 ? "BOTH" : "ALL"), inches, cm])
  end
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
  @messages.vote(user_hash(e.msg), 1) or @irc.msg(e.nick, "SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
end

def downvote(e, params)
  @messages.vote(user_hash(e.msg), -1) or @irc.msg(e.nick, "SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
end

# Reports the last message's score
def score(e, params)
  if !@messages.last
    @irc.msg(e.channel || e.nick, "NO LAST MESSAGE OR IT WAS DELETED BY !DOWNVOTE")
    return
  end

  @irc.msg(e.channel || e.nick, "#{@messages.last.text}: #{@messages.last.score}, SUBMITTED BY #{@messages.last.author}")
end

def help(e, params)
  commands = VALID_COMMANDS.collect {|cmd| "!" + cmd.to_s.upcase}.join(" ")
  target = e.channel || e.nick
  send = lambda{|msg| @irc.msg(target, msg)}

  # Show generic help if no specific command is given
  if params.empty?
    send.call "I HAVE COMMANDS AND THEY ARE: #{commands}"
    return
  end

  # Two or more params == bad news
  if params.length > 1
    send.call "WTF ARE YOU DUMB?  I OFFER HELP FOR ONE COMMAND AT A TIME JERKFACE"
    return
  end

  # Only allow valid commands' sub-help
  unless VALID_COMMANDS.include?(params.first.downcase.to_sym)
    send.call "!#{params.first} IS NOT A COMMAND YOU TWIT"
    return
  end

  case params.first
    when "DWALL" then send.call "!DWALL: SHOW EVERYBODY'S RANK, EVEN THE WORTHLESS FOLK"
    when "DONGWINNERS" then send.call "!DONGWINNERS [NUMBER]: SHOW THE PEOPLE WHO " +
                                      "FUCKING MATTER, BY DEFAULT DOES THE " +
                                      "TOP 2 FOR THE DAY... JUST LIKE A SLOW NIGHT FOR YERMOM"
    when "DONGRANKME" then send.call "!DONGRANKME: SHOW YOUR RELATIVE WORTH"
    when "SIZE" then      send.call "!SIZE [USERNAME]: GIVES YOU THE ONLY THING THAT MATTERS ABOUT SOMEBODY: SIZE"
    when "SIZEME" then    send.call "!SIZEME: TELLS YOU IF YOU ARE WORTH ANYTHING TO SOCIETY"
    when "HELP" then      send.call "OH WOW YOU ARE SO META I AM SO IMPRESSED WE SHOULD GO HAVE SEX NOW"
    when "DONGME" then    send.call "!DONGME: SHOWS HOW MUCH OF A MAN YOU ARE"
    when "REDONGME" then  send.call "!REDONGME: LETS YOU TRY TO MAKE YOURSELF INTO MORE OF A MAN BUT WITH DANGER RISK!"
    when "OMAKASE" then   send.call "!OMAKASE [TOOLNAME]: MAKES TOOLS REALLY GREAT INSTEAD OF GIANT PILES OF POORLY-ARCHITECTED BULLSHIT"
    else                  send.call "!#{params.first}: DOES SOMETHING AWESOME"
  end
end

# Reports anybody's current dong size
def size(e, params)
  if (params.empty?)
    help(e, ["SIZE"])
    return
  end

  name = params.first.upcase
  msg = nil

  for key, data in @size_data
    datanick = data[:nick].upcase
    if datanick == name
     cm = data[:size] / 2.0
     inches = cm / 2.54
     fmt = name == e.nick.upcase ? "HEY %s YOUR DONG IS" : "%s'S DONG IS"
     msg = "#{fmt} %0.1f INCHES (%0.1f CM)" % [name, inches, cm]
     break
    end
  end

  msg ||= "ONOES THERE IS NO DONG FOR #{name}"

  @irc.msg(e.channel || e.nick, msg.upcase)
end

# Reports users's current dong size
def sizeme(e, params)
  size(e, [e.nick])
end

def dongrankme(e, params)
  uhash = user_hash(e.msg)
  user_size_data = @size_data[uhash]
  if !user_size_data
    @irc.msg(e.channel || e.nick, "YOU DON'T HAVE A DONG DUMBASS")
    return
  end

  rank = size_to_rank(user_size_data[:size])
  @irc.msg(e.channel || e.nick, "YOU ARE CURRENTLY RANKED %s!" % placetext(rank))
end

def dwall(e, params)
  dongwinners(e, [10])
end

def dongwinners(e, params)
  winners_list = rank_by_size
  places = params.first.to_i
  places = 2 if places > 10 || places < 2

  output = []
  index = 0
  for winners in winners_list[0..places-1]
    nametext = userlist_text(winners.collect {|user| user[:nick]}).upcase
    size = winners.first[:size]
    cm = size / 2.0
    inches = cm / 2.54

    output.push "IN #{PLACES[index]} PLACE WE HAVE #{nametext}"
    index += 1
  end

  @irc.msg(e.channel || e.nick, output.join("; "))
end

# Reloads the ignores list if the appropriate credentials are used
def refresh_ignores(e, params)
  return unless params.first == @password
  load_ignore_list
end

# Mocks Rails and particularly DHH
def omakase(e, params)
    tool = params.empty? ? "SUPERLOUD" : params.join(" ").upcase
    if tool == "RAILS" || tool == "RUBY ON RAILS"
      response = "#{tool} IS OMAKASE TIMES INFINITY AND NOT AT ALL A PROJECT THAT'S SLOWLY TURNED INTO A NIGHTMARE OF SHITTY OPINIONATED NON-ARCHITECTURE"
    else
      response = "#{tool} IS OMAKASE"
    end
    @irc.msg(e.channel || e.nick, response)
end

#####
# Command helpers
#####

# Takes an event message and returns a semi-unique hash number representing the user + host
def user_hash(message)
  return Digest::MD5.hexdigest(message.user).to_i(16)
end

# Computes and seeds RNG with the given event's user hash combined with today's date and optional
# added value, yields to the passed-in block, and then resets the seed to the old value.
def user_seed(user_hash, add)
  old_seed = srand(user_hash + Date.today.strftime("%Y%m%d").to_i + add)
  yield
  srand(old_seed)
end

# Returns the penalty for the given number of rolls.  We want to encourage trying a second or third
# time, but discourage rolling 5 or 10 times hoping for a lucky roll to balance out the penalty.
def roll_penalty(count)
  vals = [0, -1, -3, -6, -9, -12, -18, -24, -36, -48, -60, -72, -84, -96, -108]

  return vals[count] || vals.last
end

# Computes size for a given event message.  Stores data into list of sizes for the day.
def compute_size(e)
  user_hash = user_hash(e.msg)
  mulligans = @redongs[user_hash]

  if !@ssl_users[e.msg.nick]
    @irc.log.debug "+2 redong penalty for non-ssl users"
    mulligans += 2
  end

  # Modify size by an increasing value - more redongs means more and more loss
  size_modifier = roll_penalty(mulligans)

  # Get size by using daily seed
  size = 0
  user_seed(user_hash, mulligans * 53) do
    # Is it MICROPENIS MONDAY?!?
    if Time.now.wday == 1
      size = [2, microdong_size + size_modifier].max
    else
      size = [2, fair_dong_size + size_modifier].max
    end
  end

  @size_data[user_hash] = {:size => size, :nick => e.nick, :hash => user_hash}

  return size
end

# Returns a hash of size => userdata
def users_by_size
  map = {}
  for user in @size_data.values
    size = user[:size]
    map[size] ||= []
    map[size].push(user)
  end
  return map
end

def placetext(place)
  ptext = PLACES[place - 1]
  return ptext || "NUMBER #{place.to_i}"
end

# Returns a hash of size => rank
def size_to_rank(size)
  rank = 1
  map = {}
  for mapped_size in users_by_size.keys.sort.reverse
    return rank if size == mapped_size
    rank += 1
  end
end

# Returns a list of usernames and sizes, sorted by size.  Each element in the
# returned array is a hash containing an array of users and their size.
def rank_by_size
  ranked_users = []
  for size in users_by_size.keys.sort.reverse
    ranked_users.push(users_by_size[size])
  end

  return ranked_users
end

# Helper for displaying winner text ("NERDMASTER", "NERDMASTER AND JON", etc)
def userlist_text(userlist)
  userlist.sort!
  case userlist.length
    when 1 then return userlist.first
    when 2 then return userlist.join(" AND ")
    else        return userlist[0..-2].join(", ") + ", AND #{userlist.last}"
  end
end

# Returns a value in 1/2cm units of a randomized, normalized microdong length.
# This is a simplified copy of the normal code except that we roll 3d5 instead
# of 3d15.  This makes the average around 1/2 normal, while the maximum is
# about 1/3rd the normal maximum.
def microdong_size
  average = 28
  roll = Dice.roll(3, 5) - 22
  percent = (100 + roll * 3) / 100.0
  return (percent * average).to_i
end

# Returns a value in 1/2cm units of a randomized, normalized dong length
def fair_dong_size
  # This is the "adjusted" average now and just gives us our base for the final range
  average = 28

  # Now for the fun:
  # * Pick a normalized number from -19 to +23, normalizing so closer to 2 is more common
  # * If <= 0, user is average or smaller
  #   * Percent is: (100 + number * 3) / 100.0
  # * If > 0, user is larger than average:
  #   * Percent is: (100 + number * 6) / 100.0
  # * Multiply percent by our average, giving us final size in 1/2cm units

  # 3d15 - 22, where a d15 is 1-15, gives us our -19 to +23.  Rolling multiple dice is basically
  # free normalization.
  roll = Dice.roll(3, 15) - 22

  if roll > 0
    percent = (100 + roll * 6) / 100.0
  else
    percent = (100 + roll * 3) / 100.0
  end

  return (percent * average).to_i
end

# Just keepin' the plagiarism alive, man.  At least in my version, size is always based on requester.
def send_dong(e)
  size = compute_size(e)
  @irc.msg(e.channel || e.nick, "8%sD" % ['=' * size])
end
