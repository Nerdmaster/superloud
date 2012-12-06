# Migrates messages file to a new file or format, defaulting to a YAML-to-SQLite convert
#
# Usage: ruby migrate.rb {old crappy yaml file} {new spiffy sqlite file}

require 'rubygems'
require File.dirname(__FILE__) + "/../data/messages/factory.rb"

from = ARGV.shift || "loud_messages.yml"
to = ARGV.shift || "loudness.db"

if (!from || !to)
  puts "OH COME ON YOU HAVE TO SPECIFY THE YAML FILE AND THE SQLITE FILE DUMMY"
  exit
end

if !FileTest.exist?(from)
  puts "OH COME ON THE FILE YOU SPECIFIED AS 'FROM' DOESN'T EXIST"
  exit
end

puts "OKAY GUY I WILL MIGRATE FROM #{from.inspect} TO #{to.inspect} NOW"
@old_msg = Louds::Data::Messages::Factory.create(from)
@old_msg.load
@new_msg = Louds::Data::Messages::Factory.create(to)

for text, old_message in @old_msg.instance_variable_get("@messages")
  @new_msg.add(old_message.to_hash)
end

@new_msg.serialize
