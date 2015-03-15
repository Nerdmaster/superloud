require "./rps/rps_object.rb"

RPSObject.load_rps("rps/rps.yml")

expected_win_count = nil
losses = {}
wins = {}

for object in RPSObject.objects
  defeated_objects = RPSObject.messages[object].keys

  for defeated in defeated_objects
    if !RPSObject.valid_object?(defeated)
      puts "ERROR: invalid object #{defeated.inspect} in #{object.inspect} list!"
    end

    losses[defeated] ||= []
    losses[defeated] += [object]
  end

  if defeated_objects.uniq != defeated_objects
    puts "ERROR: #{object.inspect} has at least one duped item in its list of victories"
  end

  win_count = defeated_objects.count
  expected_win_count ||= win_count
  if expected_win_count != win_count
    puts "Notice: Win count for #{object.inspect} is #{win_count}, expecting #{expected_win_count}"
  end

  wins[object] = defeated_objects
end

for object, winners in losses
  if winners.count != expected_win_count
    puts "Notice: Loss count for #{object.inspect} is #{winners.count}, expecting #{expected_win_count}"
  end

  for winner in winners
    if wins[object].include?(winner)
      puts "ERROR: #{object.inspect} both defeats and is defeated by #{winner}"
    end
  end
end
