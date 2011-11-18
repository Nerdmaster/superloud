# The RPS game code lives here

# User entered !RPS - if it's a PM we have one set of rules, channel message another
def rps(e, params)
  return rps_pm(e, params) if e.pm?
  return rps_channel(e, params)
end

def invalid_object_output(to)
  @irc.msg(to, "OBJECT MUST BE ONE OF THE FOLLOWING: %s" % RPSObject.objects.collect {|obj| obj.to_s.upcase}.join(", "))
end

def rps_pm(e, params)
  user_hash = user_hash(e.msg)
  @rps_data[user_hash] = {:nick => e.nick, :rps => RPSObject.new} unless @rps_data[user_hash]
  rps = @rps_data[user_hash][:rps]

  # User needs to specify channel and object - if either are missing, send usage info
  if params.count != 2
    @irc.msg(e.nick, "USAGE: !RPS #channel [object]")
    invalid_object_output(e.nick)
    return
  end

  (channel, object) = params

  # Channel must be valid
  unless @channel_list.include?(channel)
    @irc.msg(e.nick, "HEY DUMBFACE I AM NOT IN THAT CHANNEL I AM ONLY IN %s" % @channel_list.join(", "))
    return
  end

  # Object requested must also be valid
  object_symbol = object.downcase.to_sym

  unless RPSObject.valid_object?(object_symbol)
    @irc.msg(e.nick, "HEY DUMBFACE THAT IS AN INVALID OBJECT")
    invalid_object_output(e.nick)
    return
  end

  # Just to be a real dick - if object was valid, but not loud, user fails
  unless object.upcase == object
    @irc.msg(e.nick, "IT SOUNDED LIKE YOU REQUESTED A #{object.upcase} BUT I CAN'T QUITE HEAR YOU")
    return
  end

  # We got here?  This means our user isn't a complete idiot!
  rps.type = object_symbol
  @irc.msg(e.nick, "OKAY I WILL SET YOU UP THE BOMB ALSO THX 4 PLAYING")

  # On the other hand, the user might be a complete idiot
  if @rps_contestant[channel] && @rps_contestant[channel] == @rps_data[user_hash]
    sleep 2
    @irc.msg(e.nick, "GET THE FUCK OUT OF MY FACE THIS IS NOT FIGHT CLUB YOU CANNOT FIGHT YOURSELF")
    return
  end

  rpsname = RPSObject.config[:name].upcase

  unless @rps_contestant[channel]
    @rps_contestant[channel] = @rps_data[user_hash]
    @irc.msg(channel, "#{e.nick.upcase} HAS REGISTERED FOR #{rpsname}!  WHO IS BRAVE ENOUGH TO FIGHT???")
    return
  end

  challenger = @rps_data[user_hash]
  challengee = @rps_contestant[channel]
  @irc.msg(channel, "#{challenger[:nick].upcase} HAS REGISTERED FOR #{rpsname} TO CHALLENGE #{challengee[:nick].upcase}...")

  case challenger[:rps].fight!(challengee[:rps])
    when true
      text = "AND DEFEATS #{challengee[:nick].upcase}: %s"
    when false
      text = "BUT #{challenger[:nick].upcase} IS DEFEATED: %s"
    when nil
      text = "BUT BOTH WERE USING #{challenger[:rps].type.to_s.upcase}.  OMFG HOW GAY A TIE."
  end

  # I AM DELIBERATELY NOT UPPERCASING THE FIGHT MESSAGE IT IS TOO OBNOXIOUS OMG
  @irc.msg(channel, text % challenger[:rps].fight_message(challengee[:rps]) )

  @rps_contestant[channel] = nil
end

def rps_channel(e, params)
  @irc.msg(e.channel, "SO YOU WANT TO PLAY RPS RIGHT?  YOU GOTTA SHOW ME THE PM LOVE, BABY.  /msg #{@irc.me} !RPS [CHANNEL] [OBJECT]")
end
