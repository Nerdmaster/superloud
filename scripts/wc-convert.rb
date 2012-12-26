#!/usr/bin/env ruby
# Converts weechat logs into Messages data

require "./loudbot"
require "./data/messages"

lines = IO.readlines("log.txt")

@messages = Louds::Data::Messages::Factory.create(ARGV.shift)
@messages.load

for line in lines
  (time, user, text) = line.split(/\t/)
  text.strip!

  # Skip non-user messages
  next if user !~ /[A-Za-z]/

  # Remove special characters which designate IRC status
  user.sub!(/^[@+]/, "")

  # Ignore anything that isn't loud
  loud = false
  is_it_loud?(text) do |status, reason|
    loud = status == :loud
  end
  next unless loud

  # Ignore superloud's own messages
  next if user =~ /SUPERLOUD/

  # Ignore commands
  next if text =~ /^!/

  # Ignore messages containing bot's name
  next if text =~ /SUPERLOUD/

  @messages.add(:text => text, :author => user)
end

@messages.serialize
