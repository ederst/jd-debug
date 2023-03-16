package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/urso/sderr"
)

const localSystemJournalID = "LOCAL_SYSTEM_JOURNAL"

func openJournal(path string) (*sdjournal.Journal, error) {
	if path == localSystemJournalID || path == "" {
		j, err := sdjournal.NewJournal()
		if err != nil {
			err = sderr.Wrap(err, "failed to open local journal")
		}
		return j, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, sderr.Wrap(err, "failed to read meta data for %{path}", path)
	}

	if stat.IsDir() {
		j, err := sdjournal.NewJournalFromDir(path)
		if err != nil {
			err = sderr.Wrap(err, "failed to open journal directory %{path}", path)
		}
		return j, err
	}

	j, err := sdjournal.NewJournalFromFiles(path)
	if err != nil {
		err = sderr.Wrap(err, "failed to open journal file %{path}", path)
	}
	return j, err
}

func main() {
	j, err := openJournal("/var/log/journal")
	if err != nil {
		fmt.Printf("error creating journal: %v\n", err)
		os.Exit(1)
	}

	bootID, err := j.GetBootID()
	if err != nil {
		fmt.Printf("error getting bootID: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("BootID: %s\n", bootID)

	err = j.SeekTail()
	if err == nil {
		// funny, the SeekTail() says to use .Previous(), but .Next() is in the filebeat code
		// doesn't seem to make a difference though
		// _, err = j.Previous()
		_, err = j.Next()
	}
	// err = j.SeekHead()
	if err != nil {
		fmt.Printf("error when seeking: %v\n", err)
	}

	for {
		fmt.Println("run loop")
		c, err := j.Next()
		//c, err := j.Previous()

		if err != nil {
			fmt.Printf("error on next: %v\n", err)
			os.Exit(1)
		}

		switch {
		// error while reading next entry
		case c < 0:
			fmt.Printf("error while reading next entry %+v\n", syscall.Errno(-c))
			continue
		// no new entry, so wait
		case c == 0:
			fmt.Println("entry empty or no new entry? waiting...")
			change := j.Wait(sdjournal.IndefiniteWait)

			// log shows that it will recognice events, but will not return content for entries
			// fmt.Printf("change: %d\n", change)
			switch change {
			case sdjournal.SD_JOURNAL_NOP:
				fmt.Println("change: noop")
			case sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
				fmt.Println("change: append or vaccum")
			default:
				fmt.Println("change: unknown")
			}
			continue
		// new entries are available
		default:
		}

		fmt.Println("get entry...")
		entry, err := j.GetEntry()
		if err != nil {
			fmt.Printf("error while getting entry: %v\n", c)
			os.Exit(1)
		}

		fmt.Printf("entry: %+v", entry)
	}
}
