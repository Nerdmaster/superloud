require "sqlite3"
lib 'data/messages'

module Louds
module Data

class Messages::SQLite < Messages
  # We expect to pull at least this many messages from sqlite
  MINIMUM_RETRIEVAL_SIZE = 100

  # We limit our query to no more than 100x the minimum set
  MAXIMUM_RETRIEVAL_SIZE = MINIMUM_RETRIEVAL_SIZE * 100

  def initialize(filename)
    # TODO: Figure out a smarter way to test for the messages table
    db_exists = FileTest.exists?(filename)

    super

    @db = SQLite3::Database.new(filename)
    @db.results_as_hash = true

    # Create the db if it doesn't exist
    if !db_exists
      sql = %Q|
        BEGIN;

        CREATE TABLE messages (
          id INTEGER PRIMARY KEY,
          text TEXT,
          author TEXT,
          score INTEGER,
          views INTEGER
        );

        CREATE INDEX messages_id_idx ON messages (id);
        CREATE UNIQUE INDEX messages_text_idx ON messages (text);
        CREATE INDEX messages_author_idx ON messages (author);
        CREATE INDEX messages_score_idx ON messages (score);
        CREATE INDEX messages_views_idx ON messages (views);

        COMMIT;
      |

      @db.execute_batch(sql)
    end

    # This hash maps our messages to their ids on load, so we can update by id instead of searching
    # message text to do the update
    @message_ids = {}
  end

  # Returns true if the item exists anywhere in our database
  def exists?(text)
    # First look for the item in memory
    return true if @messages[text]

    # If not in memory, since we don't always load everything, we have to check the db
    results = @db.execute("SELECT id FROM messages WHERE text = ?", text)
    return results.length > 0
  end

  def write_data
    # Add new messages
    for message in @new_messages
      statement = "INSERT INTO messages (text, score, author, views) VALUES (?, ?, ?, ?)"
      args = [statement, message.text, message.score, message.author, message.views]
      $stderr.puts(args.inspect)
      @db.execute(*args)
      $stderr.puts "Last autoinsert id: #{@db.last_insert_row_id}"
      message.instance_variable_set("@uid", @db.last_insert_row_id)
    end

    # Update changed messages
    for message in @changed_messages
      statement = "UPDATE messages SET score = ?, views = ? WHERE id = ?"
      args = [statement, message.score, message.views, message.uid]
      $stderr.puts(args.inspect)
      @db.execute(*args)
    end
  end

  # Uses our exciting scoring logic to determine what messages to pull for a given "round" of
  # messages.  First, we pull all messages from the DB one at a time, with a minimum score
  # threshold.
  def retrieve_messages
    # First, pull messages from the DB, favoring items with fewer views.  Skip anything
    # with a score of -10 or lower.
    allresults = @db.execute(%Q|
      SELECT id AS uid, text, score, author, views
      FROM messages WHERE score > -1
      ORDER BY views
      LIMIT #{MAXIMUM_RETRIEVAL_SIZE}
    |)

    # We don't really care about order - we use "ORDER BY views" simply to
    # ensure new items get into the list in the unlikely event that we actually
    # accumulate more than 10k messages
    allresults.shuffle!

    # Just return the full set of results if we didn't pull very many beyond
    # the bare minimum
    if allresults.length <= MINIMUM_RETRIEVAL_SIZE*2
      return allresults
    end

    results = []
    if results.length < MINIMUM_RETRIEVAL_SIZE
      # We only want to pull a few over MINIMUM_RETRIEVAL_SIZE items on
      # average, so our base weight is double the ratio of minimum items to
      # total count retrieved.  This should make it very unlikely to pull
      # fewer than the minimum without having too many.
      base_chance = 3 * (MINIMUM_RETRIEVAL_SIZE.to_f / allresults.count)

      for message in allresults
        # Weight formula: the higher the view count, the higher score needs to be to get a fair shot.
        weight = base_chance * ((1+message["score"].to_f) ** 0.75 / (message["views"]+1))

        # Add extra skew in favor of zero-view items
        weight *= 2 if message["views"] == 0

        # If we decide to keep it, push this message to the back of the stack
        if rand < weight
          results.push message
        end
      end
    end

    return results
  end
end

end
end
