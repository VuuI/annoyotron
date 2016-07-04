/*
 ************************ MX ANNOYOTRON *********************
 *********** Based on hammerandchisel's airhornbot **********
 */

package main

import (
        "bytes"
        "encoding/binary"
        "flag"
        "fmt"
        "io"
        "math/rand"
        "os"
        "os/exec"
        "os/signal"
        "strconv"
        "strings"
        "time"
        "runtime"
        "regexp"
        "text/tabwriter"
        
        log "github.com/Sirupsen/logrus"
        "github.com/bwmarrin/discordgo"
        "github.com/layeh/gopus"
        )

var (
     // discordgo session
     discord *discordgo.Session
     
     
     // Map of Guild id's to *Play channels, used for queuing and rate-limiting guilds
     queues map[string]chan *Play = make(map[string]chan *Play)
     
     // Sound encoding settings
     BITRATE        = 128
     MAX_QUEUE_SIZE = 6
     
     // Owner
     OWNER string
     
     //Version
     VERSION_RELEASE = "1.0.9"
     
     //TIME Constant
     t0 = time.Now()
     
     COUNT int = 0
     
     // Shard (or -1)
     SHARDS []string = make([]string, 0)
     
     mem runtime.MemStats
     
     
     )

// Play represents an individual use of the meme sounds commands
type Play struct {
    GuildID   string
    ChannelID string
    UserID    string
    Sound     *Sound
    
    // The next play to occur after this, only used for chaining sounds like anotha
    Next *Play
    
    // If true, this was a forced play using a specific meme sound name
    Forced bool
}

type SoundCollection struct {
    Prefix    string
    Commands  []string
    Sounds    []*Sound
    ChainWith *SoundCollection
    
    soundRange int
}

// Sound represents a sound clip
type Sound struct {
    Name string
    
    // Weight adjust how likely it is this song will play, higher = more likely
    Weight int
    
    // Delay (in milliseconds) for the bot to wait before sending the disconnect request
    PartDelay int
    
    // Channel used for the encoder routine
    encodeChan chan []int16
    
    // Buffer to store encoded PCM packets
    buffer [][]byte
}
// Array of all the sounds we have
var AIRHORN *SoundCollection = &SoundCollection{
	Prefix: "airhorn",
	Commands: []string{
		"!airhorn",
	},
	Sounds: []*Sound{
		createSound("default", 1000, 250),
		createSound("reverb", 800, 250),
		createSound("spam", 800, 0),
		createSound("tripletap", 800, 250),
		createSound("fourtap", 800, 250),
		createSound("distant", 500, 250),
		createSound("echo", 500, 250),
		createSound("clownfull", 250, 250),
		createSound("clownshort", 250, 250),
		createSound("clownspam", 250, 0),
		createSound("highfartlong", 200, 250),
		createSound("highfartshort", 200, 250),
		createSound("midshort", 100, 250),
		createSound("truck", 10, 250),
		createSound("illuminati", 450, 250),
	},
}

var KHALED *SoundCollection = &SoundCollection{
	Prefix:    "another",
	ChainWith: AIRHORN,
	Commands: []string{
		"!anotha",
		"!anothaone",
	},
	Sounds: []*Sound{
		createSound("one", 1, 250),
		createSound("one_classic", 1, 250),
		createSound("one_echo", 1, 250),
	},
}

var CENA *SoundCollection = &SoundCollection{
	Prefix: "jc",
	Commands: []string{
		"!johncena",
		"!cena",
	},
	Sounds: []*Sound{
		createSound("airhorn", 1, 250),
		createSound("echo", 1, 250),
		createSound("full", 1, 250),
		createSound("jc", 1, 250),
		createSound("nameis", 1, 250),
		createSound("spam", 1, 250),
	},
}

var ETHAN *SoundCollection = &SoundCollection{
	Prefix: "ethan",
	Commands: []string{
		"!ethan",
		"!eb",
		"!ethanbradberry",
		"!h3h3",
	},
	Sounds: []*Sound{
		createSound("areyou_classic", 100, 250),
		createSound("areyou_condensed", 100, 250),
		createSound("areyou_crazy", 100, 250),
		createSound("areyou_ethan", 100, 250),
		createSound("classic", 100, 250),
		createSound("echo", 100, 250),
		createSound("high", 100, 250),
		createSound("slowandlow", 100, 250),
		createSound("cuts", 30, 250),
		createSound("beat", 30, 250),
		createSound("sodiepop", 1, 250),
	},
}

var COW *SoundCollection = &SoundCollection{
	Prefix: "cow",
	Commands: []string{
		"!stan",
		"!stanislav",
	},
	Sounds: []*Sound{
		createSound("herd", 10, 250),
		createSound("moo", 10, 250),
		createSound("x3", 1, 250),
	},
}
var DAMN *SoundCollection = &SoundCollection{
Prefix: "damn",
Commands: []string{
    "!damn",
},
    
Sounds: []*Sound{
    createSound("classic", 1000, 250),
},
}

var DEEZNUTZ *SoundCollection = &SoundCollection{
Prefix:    "deezNuts",
Commands: []string{
    "!deez",
    "!deeznutz",
},
Sounds: []*Sound{
    createSound("classic", 1000, 250),
},
}

var HITMARKER *SoundCollection = &SoundCollection{
Prefix: "hitMarker",
Commands: []string{
    "!hitmarker",
},
Sounds: []*Sound{
    createSound("classic", 200, 250),
},
}

var MMMSAY *SoundCollection = &SoundCollection{
Prefix: "mmmsay",
Commands: []string{
    "!whatcha",
    "!mmmsay",
},
Sounds: []*Sound{
    createSound("classic", 1000, 250),
},
}

var SCREAM *SoundCollection = &SoundCollection{
Prefix: "scream",
Commands: []string{
    "!wilhelm",
    "!scream",
},
Sounds: []*Sound{
    createSound("classic", 1000, 250),
},
}

var WOW *SoundCollection = &SoundCollection{
Prefix: "wow",
Commands: []string{
    "!wow",
},
Sounds: []*Sound{
    createSound("classic", 100, 250),
},
}

var TRIPLE *SoundCollection = &SoundCollection{
Prefix: "triple",
Commands: []string{
    "!ohbaby",
    "!triple",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var ILLKILLYOU *SoundCollection = &SoundCollection{
Prefix: "kill",
Commands: []string{
    "!illkillyou",
    "!killyou",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var YOUDIP *SoundCollection = &SoundCollection{
Prefix: "dip",
Commands: []string{
    "!youdip",
    "!dip",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var RUBY *SoundCollection = &SoundCollection{
Prefix: "ruby",
Commands: []string{
    "!ruby",
    "!rubymadness",
},
Sounds: []*Sound{
    createSound("classic", 100, 250),
    createSound("classic2", 100, 250),
    createSound("cheeky", 100, 250),
    createSound("scream", 100, 250),
    createSound("scream2", 100, 250),
    createSound("longscream", 100, 250),
},
}

var DEDODATED *SoundCollection = &SoundCollection{
Prefix: "dedodated",
Commands: []string{
    "!wam",
    "!dedodatedwam",
    "!dedodated",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var CLIPVIOLIN *SoundCollection = &SoundCollection{
Prefix: "clipviolin",
Commands: []string{
    "!clipviolin",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var TROMBONE *SoundCollection = &SoundCollection{
Prefix: "trombone",
Commands: []string{
    "!trombone",
	"!wahwah",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var VIOLIN *SoundCollection = &SoundCollection{
Prefix: "violin",
Commands: []string{
    "!violin",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var WEED *SoundCollection = &SoundCollection{
Prefix: "weed",
Commands: []string{
    "!weed",
    "!snoop",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var LYIN *SoundCollection = &SoundCollection{
Prefix: "lyin",
Commands: []string{
    "!lyin",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var LITTLEBOT *SoundCollection = &SoundCollection{
Prefix: "littlebot",
Commands: []string{
    "!littlebot",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var PRETTYGOOD *SoundCollection = &SoundCollection{
Prefix: "prettygood",
Commands: []string{
    "!prettygood",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var LONGSTORYSHORT *SoundCollection = &SoundCollection{
Prefix: "longstoryshort",
Commands: []string{
    "!longstoryshort",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var VAPENATION *SoundCollection = &SoundCollection{
Prefix: "vapenation",
Commands: []string{
    "!vapenation",
	"!vn",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var TINA *SoundCollection = &SoundCollection{
Prefix: "tina",
Commands: []string{
    "!fatlard",
	"!tina",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var FHRITP *SoundCollection = &SoundCollection{
Prefix: "fhritp",
Commands: []string{
    "!fhritp",
	"!fuckherrightinthepussy",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var JONTRON *SoundCollection = &SoundCollection{
Prefix: "jontron",
Commands: []string{
    "!jontron",
	"!jt",
},
Sounds: []*Sound{
    createSound("pooprat", 10, 250),
	createSound("strangerdanger", 10, 250),
	createSound("igetit", 10, 250),
	createSound("favorite", 10, 250),
	createSound("bestgame", 10, 250),
	createSound("iainthavingthatshit", 10, 250),
},
}

var FAIL *SoundCollection = &SoundCollection{
Prefix: "fail",
Commands: []string{
    "!fail",
	"!priceisright",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var EPIC *SoundCollection = &SoundCollection{
Prefix: "epic",
Commands: []string{
    "!epic",
	"!epicforthewin",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var SPAGETT *SoundCollection = &SoundCollection{
Prefix: "spagett",
Commands: []string{
    "!spagett",
	"!spaghett",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var GMM *SoundCollection = &SoundCollection{
Prefix: "gmm",
Commands: []string{
    "!gmm",
},
Sounds: []*Sound{
    createSound("cupoftea1", 10, 250),
    createSound("cupoftea2", 10, 250),
    createSound("cupoftea3", 10, 250),
    createSound("jalopy", 10, 250),
    createSound("soap", 10, 250),
},
}

var GMMTEA *SoundCollection = &SoundCollection{
Prefix: "gmm",
Commands: []string{
    "!cupoftea",
},
Sounds: []*Sound{
    createSound("cupoftea1", 10, 250),
    createSound("cupoftea2", 10, 250),
    createSound("cupoftea3", 10, 250),
},
}
var MAD *SoundCollection = &SoundCollection{
Prefix: "mad",
Commands: []string{
    "!mad",
	"!yuhefftobemad",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var BLOGE *SoundCollection = &SoundCollection{
Prefix: "bloge",
Commands: []string{
    "!bloge",
	"!questionblock",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var BANANAS *SoundCollection = &SoundCollection{
Prefix: "bananas",
Commands: []string{
    "!bananas",
	"!banana",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var ILLUMINATI *SoundCollection = &SoundCollection{
Prefix: "illuminati",
Commands: []string{
    "!illuminati",
	"!iluminati",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var RICHARD *SoundCollection = &SoundCollection{
Prefix: "richard",
Commands: []string{
    "!richard",
	"!wtfrichard",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var SPONGEBOB *SoundCollection = &SoundCollection{
Prefix: "spongebob",
Commands: []string{
    "!spongebob",
	"!sb",
},
Sounds: []*Sound{
    createSound("barnacles", 10, 250),
    createSound("steppin", 10, 250),
    createSound("steppinshort", 10, 250),
    createSound("maybe", 10, 250),
    createSound("fryers", 10, 250),
    createSound("fryersfull", 10, 250),
    createSound("leg", 10, 250),
    createSound("violinfull", 10, 250),
    createSound("violin", 10, 250),
    createSound("ugly", 10, 250),
    createSound("horse", 10, 250),
    createSound("anchovies", 10, 250),
    createSound("texas", 10, 250),
    createSound("kevin", 10, 250),
    createSound("correct", 10, 250),
    createSound("drink", 10, 250),
},
}
var MISCCRICKET *SoundCollection = &SoundCollection{
Prefix: "cricket",
Commands: []string{
    "!cricket",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
    createSound("short", 10, 250),
    createSound("single", 10, 250),
},
}
var FILTHYFRANK *SoundCollection = &SoundCollection{
Prefix: "ff",
Commands: []string{
    "!ff",
	"!filthyfrank",
},
Sounds: []*Sound{
    createSound("ravioli", 10, 250),
	createSound("hamburger", 10, 250),
	createSound("breakfast", 10, 250),
	createSound("igetpussy", 10, 250),
	createSound("poosy", 10, 250),
	createSound("heyboss", 10, 250),
},
}
var WHY *SoundCollection = &SoundCollection{
Prefix: "why",
Commands: []string{
    "!why",
	"!icantbelieveyouvedonethis",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var LMAO *SoundCollection = &SoundCollection{
Prefix: "lmao",
Commands: []string{
    "!lmao",
	"!ayylmao",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}
var DATBOI *SoundCollection = &SoundCollection{
Prefix: "datboi",
Commands: []string{
    "!datboi",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
    createSound("watchhim", 10, 250),
    createSound("beat", 10, 250),
},
}
var BABYGIRL *SoundCollection = &SoundCollection{
Prefix: "babygirl",
Commands: []string{
    "!babygirl",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
    createSound("forgive", 10, 250),
    createSound("hi", 10, 250),
	createSound("love", 10, 250),
},
}
var STOP *SoundCollection = &SoundCollection{
Prefix: "stop",
Commands: []string{
    "!stop",
	"!stopit",
	"!cummies",
    "!daddy",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var RACIST *SoundCollection = &SoundCollection{
Prefix: "racist",
Commands: []string{
    "!racist",
    "!nottoberacist",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var NORMIES *SoundCollection = &SoundCollection{
Prefix: "normies",
Commands: []string{
    "!gbp",
    "!tendies",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
},
}

var GAY *SoundCollection = &SoundCollection{
Prefix: "gay",
Commands: []string{
    "!gay",
},
Sounds: []*Sound{
    createSound("classic", 10, 250),
    createSound("faggot", 10, 250),
    createSound("dude", 10, 250),
},
}

var COLLECTIONS []*SoundCollection = []*SoundCollection{
    DAMN,
    DEEZNUTZ,
    HITMARKER,
    MMMSAY,
    SCREAM,
    WOW,
    TRIPLE,
    ILLKILLYOU,
    YOUDIP,
    RUBY,
    DEDODATED,
    CLIPVIOLIN,
    TROMBONE,
    VIOLIN,
    WEED,
    LYIN,
	LITTLEBOT,
	PRETTYGOOD,
	LONGSTORYSHORT,
	VAPENATION,
	AIRHORN,
	KHALED,
	CENA,
	ETHAN,
	COW,
	TINA,
	FHRITP,
	JONTRON,
	FAIL,
	EPIC,
	SPAGETT,
	GMM,
	GMMTEA,
	MAD,
	BLOGE,
	BANANAS,
	ILLUMINATI,
	RICHARD,
	SPONGEBOB,
	MISCCRICKET,
	FILTHYFRANK,
	WHY,
	LMAO,
	DATBOI,
	BABYGIRL,
	STOP,
	RACIST,
	NORMIES,
	GAY,
}

// Create a Sound struct
func createSound(Name string, Weight int, PartDelay int) *Sound {
    return &Sound{
    Name:       Name,
    Weight:     Weight,
    PartDelay:  PartDelay,
    encodeChan: make(chan []int16, 10),
    buffer:     make([][]byte, 0),
    }
}

func (sc *SoundCollection) Load() {
    for _, sound := range sc.Sounds {
        sc.soundRange += sound.Weight
        sound.Load(sc)
    }
}

func (s *SoundCollection) Random() *Sound {
    var (
         i      int
         number int = randomRange(0, s.soundRange)
         )
    
    for _, sound := range s.Sounds {
        i += sound.Weight
        
        if number < i {
            return sound
        }
    }
    return nil
}

// Encode reads data from ffmpeg and encodes it using gopus
func (s *Sound) Encode() {
    encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
    if err != nil {
        fmt.Println("NewEncoder Error:", err)
        return
    }
    
    encoder.SetBitrate(BITRATE * 1000)
    encoder.SetApplication(gopus.Audio)
    
    for {
        pcm, ok := <-s.encodeChan
        if !ok {
            // if chan closed, exit
            return
        }
        
        // try encoding pcm frame with Opus
        opus, err := encoder.Encode(pcm, 960, 960*2*2)
        if err != nil {
            fmt.Println("Encoding Error:", err)
            return
        }
        
        // Append the PCM frame to our buffer
        s.buffer = append(s.buffer, opus)
    }
}

// Load attempts to load and encode a sound file from disk
func (s *Sound) Load(c *SoundCollection) error {
    s.encodeChan = make(chan []int16, 10)
    defer close(s.encodeChan)
    go s.Encode()
    
    path := fmt.Sprintf("audio/%v_%v.wav", c.Prefix, s.Name)
    ffmpeg := exec.Command("ffmpeg", "-i", path, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
    
    stdout, err := ffmpeg.StdoutPipe()
    if err != nil {
        fmt.Println("StdoutPipe Error:", err)
        return err
    }
    
    err = ffmpeg.Start()
    if err != nil {
        fmt.Println("RunStart Error:", err)
        return err
    }
    
    for {
        // read data from ffmpeg stdout
        InBuf := make([]int16, 960*2)
        err = binary.Read(stdout, binary.LittleEndian, &InBuf)
        
        // If this is the end of the file, just return
        if err == io.EOF || err == io.ErrUnexpectedEOF {
            return nil
        }
        
        if err != nil {
            fmt.Println("error reading from ffmpeg stdout :", err)
            return err
        }
        
        // write pcm data to the encodeChan
        s.encodeChan <- InBuf
    }
}

// Plays this sound over the specified VoiceConnection
func (s *Sound) Play(vc *discordgo.VoiceConnection) {
    vc.Speaking(true)
    defer vc.Speaking(false)
    
    for _, buff := range s.buffer {
        vc.OpusSend <- buff
    }
}

// Attempts to find the current users voice channel inside a given guild
func getCurrentVoiceChannel(user *discordgo.User, guild *discordgo.Guild) *discordgo.Channel {
    for _, vs := range guild.VoiceStates {
        if vs.UserID == user.ID {
            channel, _ := discord.State.Channel(vs.ChannelID)
            return channel
        }
    }
    return nil
}

// Whether a guild id is in this shard
func shardContains(guildid string) bool {
    if len(SHARDS) != 0 {
        ok := false
        for _, shard := range SHARDS {
            if len(guildid) >= 5 && string(guildid[len(guildid)-5]) == shard {
                ok = true
                break
            }
        }
        return ok
    }
    return true
}

// Returns a random integer between min and max
func randomRange(min, max int) int {
    rand.Seed(time.Now().UTC().UnixNano())
    return rand.Intn(max-min) + min
}

// Prepares and enqueues a play into the ratelimit/buffer guild queue
func enqueuePlay(user *discordgo.User, guild *discordgo.Guild, coll *SoundCollection, sound *Sound) {
    // Grab the users voice channel
    channel := getCurrentVoiceChannel(user, guild)
    if channel == nil {
        log.WithFields(log.Fields{
                       "user":  user.ID,
                       "guild": guild.ID,
                       }).Warning("Failed to find channel to play sound in")
        return
    }
    
    // Create the play
    play := &Play{
    GuildID:   guild.ID,
    ChannelID: channel.ID,
    UserID:    user.ID,
    Sound:     sound,
    Forced:    true,
    }
    
    // If we didn't get passed a manual sound, generate a random one
    if play.Sound == nil {
        play.Sound = coll.Random()
        play.Forced = false
    }
    
    // If the collection is a chained one, set the next sound
    if coll.ChainWith != nil {
        play.Next = &Play{
        GuildID:   play.GuildID,
        ChannelID: play.ChannelID,
        UserID:    play.UserID,
        Sound:     coll.ChainWith.Random(),
        Forced:    play.Forced,
        }
    }
    
    // Check if we already have a connection to this guild
    //   yes, this isn't threadsafe, but its "OK" 99% of the time
    _, exists := queues[guild.ID]
    
    if exists {
        if len(queues[guild.ID]) < MAX_QUEUE_SIZE {
            queues[guild.ID] <- play
        }
    } else {
        queues[guild.ID] = make(chan *Play, MAX_QUEUE_SIZE)
        playSound(play, nil)
    }
}


// Play a sound
func playSound(play *Play, vc *discordgo.VoiceConnection) (err error) {
    log.WithFields(log.Fields{
                   "play": play,
                   }).Info("Playing sound")
    
    if vc == nil {
        vc, err = discord.ChannelVoiceJoin(play.GuildID, play.ChannelID, false, false)
        // vc.Receive = false
        if err != nil {
            log.WithFields(log.Fields{
                           "error": err,
                           }).Error("Failed to play sound")
            delete(queues, play.GuildID)
            return err
        }
    }
    
    // If we need to change channels, do that now
    if vc.ChannelID != play.ChannelID {
        vc.ChangeChannel(play.ChannelID, false, false)
        time.Sleep(time.Millisecond * 125)
    }
    
    
    // Sleep for a specified amount of time before playing the sound
    //time.Sleep(time.Millisecond * 32)
    
    // Play the sound
    play.Sound.Play(vc)
    
    // If this is chained, play the chained sound
    if play.Next != nil {
        playSound(play.Next, vc)
    }
    
    // If there is another song in the queue, recurse and play that
    if len(queues[play.GuildID]) > 0 {
        play := <-queues[play.GuildID]
        playSound(play, vc)
        return nil
    }
    
    // If the queue is empty, delete it
    time.Sleep(time.Millisecond * time.Duration(play.Sound.PartDelay))
    delete(queues, play.GuildID)
    vc.Disconnect()
    return nil
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
    log.Info("Recieved READY payload")
    var status string
    status = generateStatus()
    s.UpdateStatus(0, status)
	status = generateStatus()
    s.UpdateStatus(0, status)

}


func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
    
    if !shardContains(event.Guild.ID) {
        return
    }
    
    if event.Guild.Unavailable != nil {
        return
    }
    
    for _, channel := range event.Guild.Channels {
        if channel.ID == event.Guild.ID {
            s.ChannelMessageSend(channel.ID, "**HELLO HELLO HELLO**")
            return
        }
    }
}

func scontains(key string, options ...string) bool {
    for _, item := range options {
        if item == key {
            return true
        }
    }
    return false
}


// Handles bot operator messages, should be refactored (lmao)
func handleBotControlMessages(s *discordgo.Session, m *discordgo.MessageCreate, parts []string, g *discordgo.Guild) {
    ourShard := shardContains(g.ID)
    if len(parts) >= 3 && scontains(parts[len(parts)-2], "die") {
        shard := parts[len(parts)-1]
        if len(SHARDS) == 0 || scontains(shard, SHARDS...) {
            log.Info("Got DIE request, exiting...")
            s.ChannelMessageSend(m.ChannelID, ":ok_hand: goodbye cruel world")
            os.Exit(0)
        }
    } else if scontains(parts[len(parts)-1], "info") && ourShard {
        runtime.ReadMemStats(&mem)
        t1 := time.Now()
        d := t1.Sub(t0)
        minutesPassed := d.Minutes()
        var truncate int = int(minutesPassed) % 60
        var hoursPassed int = int(minutesPassed / 60)
        w := &tabwriter.Writer{}
        buf := &bytes.Buffer{}
        
        w.Init(buf, 0, 4, 0, ' ', 0)
        fmt.Fprintf(w, "```\n")
        fmt.Fprintf(w, "Discordgo: \t%s\n", discordgo.VERSION)
        fmt.Fprintf(w, "Go: \t%s\n", runtime.Version())
        fmt.Fprintf(w, "ver.: \t%s\n", VERSION_RELEASE)
        fmt.Fprintf(w, "Time Up: \t%v hrs. %v min.\n", hoursPassed, truncate)
        fmt.Fprintf(w, "Memory: \t%d / %d (%d total)\n", mem.Alloc, mem.Sys, mem.TotalAlloc)
        fmt.Fprintf(w, "Calls: \t%d\n", COUNT)
        fmt.Fprintf(w, "Servers: \t%d\n", len(discord.State.Ready.Guilds))
        fmt.Fprintf(w, "```\n")
        w.Flush()
        s.ChannelMessageSend(m.ChannelID, buf.String())
    } else if scontains(parts[len(parts)-1], "where") && ourShard {
        s.ChannelMessageSend(m.ChannelID,
                             fmt.Sprintf("its a me, shard %v", string(g.ID[len(g.ID)-5])))
    }else if scontains(parts[len(parts)-1], "killbot") && ourShard {
        s.ChannelMessageSend(m.ChannelID,":ok_hand: goodbye cruel world")
        os.Exit(0)
	}else if scontains(parts[len(parts)-1], "newstatus") && ourShard {
        var status string
        status = generateStatus()
        s.UpdateStatus(0, status)
		s.ChannelMessageSend(m.ChannelID, "Ayy new status set m8 :ok_hand:")
    }
    return
}

func generateCommandList() string{
    var commands string
    commands = "\n**HELLO HELLO HELLO**\n`Annoyotron version 1.0.9 Commands \n\nChangelog: Added !stop\n\n"
    commands = commands + "!damn,!deez,!hitmarker,!mmmsay,!scream,!wow,!triple,!illkillyou,!jontron,!fhritp,!tina,\n"
	commands = commands + "!littlebot,!prettygood,!longstoryshort,!vapenation,!airhorn,!gmm,!cupoftea,!spagett,!epic,!mad\n"
    commands = commands + "!dip,!ruby,!dedodated,!trombone,!violin,!weed,!lyin,!roll,!richard,!illuminati,!bananas,!questionblock,!cricket,!spongebob,!eb,!jc,!filthyfrank,!anotha,!why,!lmao,!datboi,!babygirl\n`"
    //commands = commands + "\n:ok_hand: 1 spam = 1 Michael BabyRage :ok_hand:"
    return commands
}
func generateStatus() string{
	rand.New(rand.NewSource(99))
    status := []string{
		"with this pussy",
		"with Patrick's fursuit",
		"Nazi Gainers",
		"with dat boi",
		"supreme meme creme",
		"( ͡° ͜ʖ ͡°)",
		"420 noscope",
		"with Shane's mom",
		"with Griff's BBQ sauce",
		"with Sam's bodypillow",
		"Town of Gaylem",
		"with Jacks goat",
		"with Kevin's waffleiron",
		"with John's skinflute",
		"with Mike's comics",
		"with Derrick's loud dice",
		"Eating gelato with Sam",
		"backdoor sluts 9",
		"Sluts and Butts IX",
		"Beyond Earth. LOL JK",
		"Earning POTG as Bastion"
		"Earning POTG as Hanzo"
		"Earning POTG as Mei"
		"Earning POTG as Reinhardt"
		"Earning POTG as Genji"
	}
    return status[rand.Intn(len(status))]
}


func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if len(m.Content) <= 0 || (m.Content[0] != '!' && len(m.Mentions) != 1) {
        return
    }
    
    parts := strings.Split(strings.ToLower(m.Content), " ")
    
    channel, _ := discord.State.Channel(m.ChannelID)
    if channel == nil {
        log.WithFields(log.Fields{
                       "channel": m.ChannelID,
                       "message": m.ID,
                       }).Warning("Failed to grab channel")
        return
    }
    
    guild, _ := discord.State.Guild(channel.GuildID)
    if guild == nil {
        log.WithFields(log.Fields{
                       "guild":   channel.GuildID,
                       "channel": channel,
                       "message": m.ID,
                       }).Warning("Failed to grab guild")
        return
    }
    
    // If this is a mention, it should come from the owner (otherwise we don't care)
    if len(m.Mentions) > 0 {
        if m.Mentions[0].ID == s.State.Ready.User.ID && m.Author.ID == OWNER && len(parts) > 0 {
            handleBotControlMessages(s, m, parts, guild)
        }
        return
    }
    
    // If it's not relevant to our shard, just exit
    if !shardContains(guild.ID) {
        return
    }
    
    // If !commands is sent
    if parts[0] == "!commands" {
        COUNT++
        var commands string
        commands = generateCommandList()
        s.ChannelMessageSend(channel.ID, commands)
        return
    }
	// statustest
    //if parts[0] == "!newstatus" {
    //    COUNT++
	//	s.ChannelMessageSend(channel.ID, "Ayy new status set m8 :ok_hand:")
    //    var status string
    //    status = generateStatus()
    //    s.UpdateStatus(0, status)
    //    return
   // }
	
	    if parts[0] == "!help" {
        COUNT++
        s.ChannelMessageSend(channel.ID, "there aint no helping you boy.")
        return
    }
    	
	    if parts[0] == "!residentsleeper" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/MSgZL.gif")
        return
	}
    	
	    if parts[0] == "!feelsbadman" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/Zm0UHtB.png")
        return
	}
    	
	    if parts[0] == "!smorc" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/2hltyV9.png")
        return
	}
    	
	    if parts[0] == "!kreygasm" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/wN4gmZX.jpg")
        return
	}
    	
	    if parts[0] == "!stinkycheese" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/eATluCU.png")
        return

    }
		if parts[0] == "!datboi" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/3udLX1g.gif")
    }
		if parts[0] == "!mute" {
        COUNT++
        s.ChannelMessageSend(channel.ID, "I'm sorry my memes bother you, consider joining the Bot-Free Lobby or right clicking my name in the contact list and ticking mute, alternatively you can kill yourself.")
        return
    }
		if parts[0] == "!kappa" {
        COUNT++
		s.ChannelMessageSend(channel.ID, " http://i.imgur.com/pjry738.png")
        return
    }
		
		if parts[0] == "!biblethump" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/MS9WtiO.png")
        return
    }
		
		if parts[0] == "!babyrage" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/wIn6Hqn.png")
        return
    }
		
		if parts[0] == "!elegiggle" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/Eqcp3Tp.jpg")
        return
    }
		
		if parts[0] == "!4head" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/9B4uqbD.jpg")
        return
    }
	
		if parts[0] == "!pogchamp" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/DixUK9r.jpg")
        return
    }
	
		if parts[0] == "!failfish" {
        COUNT++
        s.ChannelMessageSend(channel.ID, " http://i.imgur.com/potJY8t.png")
        return
    }
	
    if parts[0] == "!roll" {
        COUNT++
        re := regexp.MustCompile("^[0-9]*$")
        if len(parts) == 1 {
            var num = randomRange(1, 20)
            s.ChannelMessageSend(channel.ID, "```Rolling a dank ass d20```")
            time.Sleep(time.Millisecond * 100)
            s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%v```", num))
            return
        }else{
            var amt int = 1
            var splitD = strings.Split(parts[1], "d")
            //if a command like 2d6
            if (re.MatchString(splitD[0])){//checking if [1] is a num
                amt,_ = strconv.Atoi(splitD[0])
                if amt > 5 && m.Author.ID != OWNER{//Allows the owner to be a spammy jerk
                    s.ChannelMessageSend(channel.ID, "```Whoa there buddy, only 5 at a time```")
                    return
                }
            }
            
            if(splitD[0] == parts[1]){
                s.ChannelMessageSend(channel.ID, "```Invalid entry, try 'd20' or 'd6'```")
                return
            }

            if re.MatchString(splitD[1]){//if [1] is not a num
            }else{
                s.ChannelMessageSend(channel.ID, "```Invalid entry, try 'd20' or 'd6'```")
                return
            }
            
            if splitD[1] == ""{
                s.ChannelMessageSend(channel.ID, "```Invalid entry, try 'd20' or 'd6'```")
                return
            }
            var max int
            max,_ = strconv.Atoi(splitD[1])
            var num int
            if amt == 0 {
                amt++
            }
            for i:=0; i < amt; i++ {
                num = randomRange(1, max + 1)
                s.ChannelMessageSend(channel.ID, fmt.Sprintf("```Rolling d%v```", max))
                time.Sleep(time.Millisecond * 50)
                s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%v```", num))
            }
            return
        }
        
    }
    
    
    
    // Find the collection for the command we got
    for _, coll := range COLLECTIONS {
        if scontains(parts[0], coll.Commands...) {
            
            // If they passed a specific sound effect, find and select that (otherwise play nothing)
            var sound *Sound
            if len(parts) > 1 {
                for _, s := range coll.Sounds {
                    if parts[1] == s.Name {
                        sound = s
                    }
                }
                
                if sound == nil {
                    return
                }
            }
            COUNT++
            go enqueuePlay(m.Author, guild, coll, sound)
            return
        }
    }
}

func main() {
    var (
         Token = flag.String("t", "", "Discord Authentication Token")
         Shard = flag.String("s", "", "Integers to shard by")
         Owner = flag.String("o", "", "Owner ID")
         err   error
         )
    flag.Parse()
    if *Owner != "" {
        OWNER = *Owner
    }
    
    // Make sure shard is either empty, or an integer
    if *Shard != "" {
        SHARDS = strings.Split(*Shard, ",")
        
        for _, shard := range SHARDS {
            if _, err := strconv.Atoi(shard); err != nil {
                log.WithFields(log.Fields{
                               "shard": shard,
                               "error": err,
                               }).Fatal("Invalid Shard")
                return
            }
        }
    }
    
    // Preload all the sounds
    log.Info("Preloading sounds...")
    for _, coll := range COLLECTIONS {
        coll.Load()
    }
    
    
    // Create a discord session
    log.Info("Starting discord session...")
    discord, err = discordgo.New(*Token)
    if err != nil {
        log.WithFields(log.Fields{
                       "error": err,
                       }).Fatal("Failed to create discord session")
        return
    }
    
    discord.AddHandler(onReady)
    discord.AddHandler(onGuildCreate)
    discord.AddHandler(onMessageCreate)
    
    err = discord.Open()
    if err != nil {
        log.WithFields(log.Fields{
                       "error": err,
                       }).Fatal("Failed to create discord websocket connection")
        return
    }
    
    // We're running!
    log.Info("MEMEBOT is ready to autist it up.")
    
    // Wait for a signal to quit
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill)
    <-c
}
