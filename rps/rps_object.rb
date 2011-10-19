require "yaml"

class RPSObject
  attr_reader :wins, :losses, :type

  def self.load_rps(filename)
    @@config = YAML.load_file(filename)
    raise "Invalid file" unless @@config.is_a?(Hash)

    @@messages = config[:messages]
    @@objects = @@messages.keys
  end

  def self.config; return @@config; end
  def self.messages; return @@messages; end
  def self.objects; return @@objects; end

  def self.valid_object?(obj)
    return @@objects.include?(obj.to_sym)
  end

  def initialize
    @wins = 0
    @losses = 0
  end

  def valid_object?
    return RPSObject.valid_object?(@type)
  end

  # Sets @type if it's valid
  def type=(val)
    object = val.to_sym
    @type = object if RPSObject.valid_object?(object)
  end

  def fight(rps)
    raise "Invalid object for challenger" unless valid_object?
    raise "Invalid object for challengee" unless rps.valid_object?

    return nil if @type == rps.type

    if @@messages[@type][rps.type]
      return true
    end

    return false
  end

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
