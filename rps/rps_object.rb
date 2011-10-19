require "yaml"

# This is kind of a rock-paper-scissors engine.  It's kind of lame, but extremely configurable.
# Call `RPSObject.load_rps("some.file.yml")` to set up the RPS rules.  The included rps.yml-dist
# should give you an idea how the configuration should be formatted.
#
# Create a new RPSObject, set its type, and you can tell it to fight another RPSObject.
class RPSObject
  attr_reader :wins, :losses, :type

  # Loads class-wide configuration for RPS engine
  def self.load_rps(filename)
    @@config = YAML.load_file(filename)
    raise "Invalid file" unless @@config.is_a?(Hash)

    @@messages = config[:messages]
    @@objects = @@messages.keys
  end

  def self.config; return @@config; end
  def self.messages; return @@messages; end
  def self.objects; return @@objects; end

  # Checks the given type for validity
  def self.valid_object?(obj)
    return @@objects.include?(obj.to_sym)
  end

  # Inits win/loss counters to zero
  def initialize
    @wins = 0
    @losses = 0
  end

  # Checks to see if @type is valid
  def valid_object?
    return RPSObject.valid_object?(@type)
  end

  # Sets @type if it's valid
  def type=(val)
    object = val.to_sym
    @type = object if RPSObject.valid_object?(object)
  end

  # Computes winner between self and passed-in object.  True means I win (self), false means you
  # win (passed-in object) and nil means tie.
  #
  # We know who the winner is based entirely on the message - if there's a message defined in
  # @@messaegs[self.type][rps.type], I win.  Look at the yaml config to see why this works.
  def fight(rps)
    raise "Invalid object for challenger" unless valid_object?
    raise "Invalid object for challengee" unless rps.valid_object?

    return nil if @type == rps.type

    if @@messages[@type][rps.type]
      return true
    end

    return false
  end

  # Runs #fight above, but also increments wins/losses for both objects
  def fight!(rps)
    if true == fight(rps)
      self.won
      rps.lost
      return true
    elsif false == fight(rps)
      self.lost
      rps.won
      return false
    end

    return nil
  end

  def fight_message(rps)
    return "TIE!" if @type == rps.type
    return @@messages[@type][rps.type] || @@messages[rps.type][@type]
  end

  def won
    @wins += 1
  end

  def lost
    @losses += 1
  end
end
