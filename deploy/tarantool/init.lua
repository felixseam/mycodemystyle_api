#! /usr/bin/tarantool

-- This is default tarantool initialization file
-- with easy to use configuration examples including
-- replication, sharding and all major features
-- Complete documentation available in:  http://tarantool.org/doc/
--
-- To start this instance please run `systemctl start tarantool@example` or
-- use init scripts provided by binary packages.
-- To connect to the instance, use "sudo tarantoolctl enter example"
-- Features:
-- 1. Database configuration
-- 2. Binary logging and snapshots
-- 3. Replication
-- 4. Automatinc sharding
-- 5. Message queue
-- 6. Data expiration
-----------------
-- Configuration
-----------------
require('compat').sql_seq_scan_default = 'old'
box.cfg {

    ------------------------
    -- Network configuration
    ------------------------

    -- The read/write data port number or URI
    -- Has no default value, so must be specified if
    -- connections will occur from remote clients
    -- that do not use “admin address”
    listen = 3301,
    work_dir = '.data/tarantool',
    -- listen = '*:3301';

    -- The server is considered to be a Tarantool replica
    -- it will try to connect to the master
    -- which replication_source specifies with a URI
    -- for example konstantin:secret_password@tarantool.org:3301
    -- by default username is "guest"
    -- replication_source="127.0.0.1:3102";

    -- The server will sleep for io_collect_interval seconds
    -- between iterations of the event loop
    io_collect_interval = nil,

    -- The size of the read-ahead buffer associated with a client connection
    readahead = 10 * 1024 * 1024,

    --- Develoepr set for Fiber:
    net_msg_max = 1048576,
    ----------------------
    -- Memtx configuration
    ----------------------

    -- An absolute path to directory where snapshot (.snap) files are stored.
    -- If not specified, defaults to /var/lib/tarantool/INSTANCE
    -- memtx_dir = nil;

    -- How much memory Memtx engine allocates
    -- to actually store tuples, in bytes.
    -- memtx_memory = 128 * 1024 * 1024;
    memtx_memory = 1024 * 1024 * 1024,

    -- Size of the smallest allocation unit, in bytes.
    -- It can be tuned up if most of the tuples are not so small
    memtx_min_tuple_size = 16,

    -- Size of the largest allocation unit, in bytes.
    -- It can be tuned up if it is necessary to store large tuples
    memtx_max_tuple_size = 10 * 1024 * 1024, -- 10Mb
    vinyl_max_tuple_size = 10 * 1024 * 1024, -- 10Mb

    ----------------------
    -- Vinyl configuration
    ----------------------

    -- An absolute path to directory where Vinyl files are stored.
    -- If not specified, defaults to /var/lib/tarantool/INSTANCE
    -- vinyl_dir = nil;

    -- How much memory Vinyl engine can use for in-memory level, in bytes.
    vinyl_memory = 128 * 1024 * 1024, -- 128 mb

    -- How much memory Vinyl engine can use for caches, in bytes.
    vinyl_cache = 64 * 1024 * 1024, -- 64 mb

    -- The maximum number of background workers for compaction.
    vinyl_write_threads = 2,

    ------------------------------
    -- Binary logging and recovery
    ------------------------------

    -- An absolute path to directory where write-ahead log (.xlog) files are
    -- stored. If not specified, defaults to /var/lib/tarantool/INSTANCE
    -- wal_dir = nil;

    -- Specify fiber-WAL-disk synchronization mode as:
    -- "none": write-ahead log is not maintained;
    -- "write": fibers wait for their data to be written to the write-ahead log;
    -- "fsync": fibers wait for their data, fsync follows each write;
    --    wal_mode = "none";
    wal_mode = "write",

    -- The maximal size of a single write-ahead log file
    wal_max_size = 256 * 1024 * 1024,

    -- The interval between actions by the snapshot daemon, in seconds
    checkpoint_interval = 60 * 60, -- one hour

    -- The maximum number of snapshots that the snapshot daemon maintans
    checkpoint_count = 6,

    -- Reduce the throttling effect of box.snapshot() on
    -- INSERT/UPDATE/DELETE performance by setting a limit
    -- on how many megabytes per second it can write to disk
    snap_io_rate_limit = nil,

    -- Don't abort recovery if there is an error while reading
    -- files from the disk at server start.
    force_recovery = true,

    ----------
    -- Logging
    ----------

    -- How verbose the logging is. There are six log verbosity classes:
    -- 1 – SYSERROR
    -- 2 – ERROR
    -- 3 – CRITICAL
    -- 4 – WARNING
    -- 5 – INFO
    -- 6 – DEBUG
    log_level = 5,
    -- log = "tarantool.log";

    -- By default, the log is sent to /var/log/tarantool/INSTANCE.log
    -- If logger is specified, the log is sent to the file named in the string
    --     log = "tarantool.log",
    --     wal_dir = './db/wal',
    --     memtx_dir = './db/memtx',
    --     vinyl_dir = './db/vinyl',
    -- work_dir = './work',

    -- If true, tarantool does not block on the log file descriptor
    -- when it’s not ready for write, and drops the message instead
    -- log_nonblock = true;

    -- If processing a request takes longer than
    -- the given value (in seconds), warn about it in the log
    too_long_threshold = 0.5,

    -- Inject the given string into server process title
    -- custom_proc_title = 'example';
    background = false,
    -- pid_file = 'rust.pid';
    memtx_use_mvcc_engine = true
}

local function bootstrap()
    box.schema.user.create('felix', {password = '122002'})
    box.schema.user.grant('felix', 'read,write,execute,create,alter,drop',
                          'universe')
    -- box.schema.user.grant('guest', 'read,write,execute', 'universe')
    -- print("Granted permissions to guest user.")
end
box.once('grants2', bootstrap)

json = require('json')
fiber = require('fiber')
uuid = require("uuid")
datetime = require("datetime")
decimal = require('decimal')
console = require("console");
log = require('log')

-- define the to_char function
if box.schema.func.exists('to_char') then
    box.schema.func.drop('to_char')
end
function to_char(datetime, format)
    -- convert the datetime value to a string
    local datetime_str = tostring(datetime)
    -- extract the year, month, day, hour, minute, second and millisecond parts
    local year = string.sub(datetime_str, 1, 4)
    local month = string.sub(datetime_str, 6, 7)
    local day = string.sub(datetime_str, 9, 10)
    local hour = string.sub(datetime_str, 12, 13)
    local minute = string.sub(datetime_str, 15, 16)
    local second = string.sub(datetime_str, 18, 19)
    local millisecond = string.sub(datetime_str, 21)
    -- replace the format placeholders with the corresponding parts
    local formatted_str = format
    formatted_str = string.gsub(formatted_str, "yyyy", year)
    formatted_str = string.gsub(formatted_str, "mm", month)
    formatted_str = string.gsub(formatted_str, "dd", day)
    formatted_str = string.gsub(formatted_str, "hh", hour)
    formatted_str = string.gsub(formatted_str, "mi", minute)
    formatted_str = string.gsub(formatted_str, "ss", second)
    formatted_str = string.gsub(formatted_str, "ms", millisecond)
    -- return the formatted string as a tuple
    return formatted_str
end

-- check if the function already exists
if box.schema.func.exists('to_char') then box.schema.func.drop('to_char'); end
if not box.schema.func.exists('to_char') then
    -- make the function callable from SQL
    box.schema.func.create('to_char', {
        language = 'LUA',
        param_list = {'datetime', 'string'},
        returns = 'string',
        exports = {'SQL'}
    })
end

-- define the unflatten_sql function
if box.schema.func.exists('unflatten_sql') then
    box.schema.func.drop('unflatten_sql');
end
function unflatten_sql(sql_res)
    local objects = {}

    for i, row in ipairs(sql_res.rows) do
        local obj = {}
        for j, col in ipairs(sql_res.metadata) do obj[col.name] = row[j] end

        objects[i] = obj
    end
    -- log.info(objects)
    return objects
end

-- check if the function already exists
if box.schema.func.exists('unflatten_sql') then
    box.schema.func.drop('unflatten_sql');
end

function get_seq_current(seq_name)
    local sequence = box.sequence[seq_name]
    local current_value = sequence:current()
    return current_value + 1
end

-- check if the function already exists
if box.schema.func.exists('get_seq_current') then
    box.schema.func.drop('get_seq_current')
end

if not box.schema.func.exists('get_seq_current') then
    -- make the function callable from SQL
    box.schema.func.create('get_seq_current', {
        language = 'LUA',
        param_list = {'string'},
        returns = 'unsigned',
        exports = {'SQL'}
    })
end

function get_seq_next(seq_name)
    local sequence = box.sequence[seq_name]
    return sequence:next() -- This is atomic and thread-safe
end

-- check if the function already exists
if box.schema.func.exists('get_seq_next') then
    box.schema.func.drop('get_seq_next')
end

if not box.schema.func.exists('get_seq_next') then
    -- make the function callable from SQL
    box.schema.func.create('get_seq_next', {
        language = 'LUA',
        param_list = {'string'},
        returns = 'unsigned',
        exports = {'SQL'}
    })
end

-- Set New value seuquence
function set_seq_val(seq_name)
    local sequence = box.sequence[seq_name]
    local next_value = sequence:next()
    if next_value ~= nil then
        local current_value = sequence:current()
        return current_value
    else
        return 0
    end
end

-- check if the function already exists
if box.schema.func.exists('set_seq_val') then box.schema.func
    .drop('set_seq_val') end

if not box.schema.func.exists('set_seq_val') then
    -- make the function callable from SQL
    box.schema.func.create('set_seq_val', {
        language = 'LUA',
        param_list = {'string'},
        returns = 'unsigned',
        exports = {'SQL'}
    })
end


--- Cache system
----------------------

local cache = box.schema.space.create('cache_entries', {
    if_not_exists = true,
    engine = 'memtx',
})

cache:format({
    { name = 'key',       type = 'string' },
    { name = 'value',     type = 'string' },
    { name = 'expires_at', type = 'unsigned' },
})

cache:create_index('primary', {
    type = 'TREE',
    unique = true,
    parts = {
        { field = 'key', type = 'string' },
    },
    if_not_exists = true,
})

cache:create_index('expires_at', {
    type = 'TREE',
    unique = false,
    parts = {
        { field = 'expires_at', type = 'unsigned' },
    },
    if_not_exists = true,
})

-- Cache expiration worker
----------------------------

local fiber = require('fiber')

local function cleanup_expired_cache()
    local cache = box.space.cache_entries

    while true do
        local now = os.time()
        local expired = cache.index.expires_at:select(
            { 0 },
            { iterator = 'GT', limit = 500 }
        )

        for _, tuple in ipairs(expired) do
            if tuple.expires_at > now then
                break
            end

            if tuple.expires_at > 0 then
                cache:delete({ tuple.key })
            end
        end

        fiber.sleep(#expired > 0 and 0.01 or 1)
    end
end

fiber.create(cleanup_expired_cache)
