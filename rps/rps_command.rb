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

  # User needs to specify channel and object - if either are missing, send usage info
  if params.count != 2
    @irc.msg(e.nick, "USAGE: !RPS #channel [object]")
    invalid_object_output(e.nick)
    return
  end

  (channel, object) = params

  unless @channel_list.include?(channel)
    @irc.msg(e.nick, "HEY DUMBFACE I AM NOT IN THAT CHANNEL I AM ONLY IN %s" % @channel_list.join(", "))
    return
  end

  unless RPSObject.valid_object?(params.first.downcase)
    @irc.msg(e.nick, "HEY DUMBFACE THAT IS AN INVALID OBJECT")
    invalid_object_output(e.nick)
    return
  end
end

def rps_channel(e, params)
  @irc.msg(e.channel, "I AM REALLY SO SORRY BUT THIS STILL ISN'T FUCKING READY OKAY")
end
