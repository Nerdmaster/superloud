module Louds
module Data

class Message
  attr_reader :author, :views, :score, :text
  attr_accessor :container

  # Loads a message's attributes from a hash
  def self.from_hash(hsh)
    msg = Message.new(hsh[:text], hsh[:author], hsh[:score], hsh[:views])

    return msg
  end

  def initialize(text, author, score = 1, views = 0)
    @text = text
    @author = author
    @views = views
    @score = score
  end

  def view!
    @views += 1
    @container.dirty!
  end

  def upvote!
    @score += 1
    @container.dirty!
  end

  def downvote!
    @score -= 1
    @container.dirty!
  end

  # Converts all important attributes to a hash of data, primarily to ease exporting
  def to_hash
    return { :author => @author, :views => @views, :score => @score, :text => @text }
  end
end

end
end
