package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/utils"
	"github.com/pkg/errors"
)

// PorgASCIIArt is the ascii art for the porg
var PorgASCIIArt = `
                                          '.:/+ooossoo+/:-'
                                     ':+ydNMMMMMMMMMMMMMMMNmyo:'
                                   '.--.''.:ohNMMMMMMMMNho:.''..'
                                        -o+.  ':sNMMms-'  .--
                                 '+o     .mNo    '::'   :dNh'    '+-
                                :mMo      dMM-         .NMMs      hNs
                               -NMMNs:--/hMMM+         .MMMNo-''-sMMMs
                          -'   sMMMMMMNNMMMMM:          hMMMMNNNNMMMMN
                         -y    /MMMMMMMMMMMNs           .mMMMMMMMMMMMd     :
                        .mN.    oNMMMMMMMms-             .yNMMMMMMMNy.    -N:
                        hMMm+'   ./syys+-'    -//   ':/-   ./syyys/.    'oNMm'
                       /MMMMMms:.              '.    '''             ./ymMMMMs
                       mMMMMMMMMNo                 '               'sNMMMMMMMN.
                      -MMMMMMMMN+'             :shmmmh+.            'oNMMMMMMMo
                      /MMMMMMMd-             :dds+///sdm/             :mMMMMMMm
                      sMMMMMMd.             :m+'      ':hs'            -mMMMMMM+
                    .hMMMMMMM-             :h:'         'oh.            /MMMMMMN/
                   /mMMMMMMMN'           ./+' .://:::::.  /d/           .MMMMMMMN/
                 'sMMMMMMMMMM/       '.-:-'     '....'     .so-'        oMMMMMMMMN:
                .hMMMMMMMMMMMNs-'     ''                     '--     '-yNMMMMMMMMMN-
               -dMMMMMMMMMMMMMMNs'                                  'yNMMMMMMMMMMMMm.
              -mMMMMMMMMMMMMMMNo'                                    'sMMMMMMMMMMMMMd'
             :NMMMMMMMMMMMMMMm:                                        /NMMMMMMMMMMMMy
            -NMMMMMMMMMMMMMMd.                                        ' -mMMMMMMMMMMMMo
           .mMMMMMMMMMMMMMMd-/o'                                      .o::mMMMMMMMMMMMN/
          'dMMMMMMMMMMMMMMMhmm.                                        -mdhMMMMMMMMMMMMm.
          yMMMMMMMMMMMMMMMMMN-                                          :NMMMMMMMMMMMMMMh
         /MMMMMMMMMMMMMMMMMN:                                            +MMMMMMMMMMMMMMM+
        'mMMMMMMMMMMMMMMMMMo                                              sMMMMMMMMMMMMMMN.
        oMMMMMMMMMMMMMMMMMh                                               'dMMMMMMMMMMMMMMy
       'mMMMMMMMMMMMMMMMMN.                                                :MMMMMMMMMMMMMMM-
       :MMMMMMMMMMMMMMMMMo                                                  yMMMMMMMMMMMMMMy
       sMMMMMMMMMMMMMMMMN'                                                  -MMMMMMMMMMMMMMN'
       dMMMMMMMMMMMMMMMMh                                                    mMMMMMMMMMMMMMM-
       mMMMMMMMMMMMMMMMMo                                                    yMMMMMMMMMMMMMM-
       mMMMMMMMMMMMMMMMMo                                                    sMMMMMMMMMMMMMM.
       hMMMMMMMMMMMMMMMMs                                                    hMMMMMMMMMMMMMN
       oMMMMMMMMMMMMMMMMy                                                    dMMMMMMMMMMMMMy
       .MMMMMMMMMMMMMMMMd                                                    NMMMMMMMMMMMMM-
        yMMMMMMMMMMMMMMMM'                                                  .MMMMMMMMMMMMMh
        .NMMMMMMMMMMMMMMM/                                                  oMMMMMMMMMMMMN-
         :NMMMMMMMMMMMMMMh                                                  mMMMMMMMMMMMMo
          /NMMMMMMMMMMMMMM-                                                /MMMMMMMMMMMMd'
           :NMMMMMMMMMMMMMh                                               'mMMMMMMMMMMMm.
            .hMMMMMMMMMMMMM/                                              oMMMMMMMMMMMN-
              +mMMMMMMMMMMMN-                                            :MMMMMMMMMMMm-
               'oNMMMMMMMMMMm.                                          -NMMMMMMMMMMd-
                 .omNmh+:hNMMm-                                        :NNsmMMMMMMMy'
                   '.     -smMN+                                     'oNh- 'sNMMNh:
                            ':yNh-                                  -hh:     .:-'
                               ':o/'                              '/+.
                                   '                              '

`

func restartDock() error {
	utils.Indent(log.Info, 2)("restarting Dock")
	if _, err := utils.RunCommand(context.Background(), "killall", "Dock"); err != nil {
		return errors.Wrap(err, "killing Dock process failed")
	}
	// let system settle
	time.Sleep(2 * time.Second)
	return nil
}

func removeOldDatabaseFiles(dbpath string) error {

	paths := []string{
		filepath.Join(dbpath, "db"),
		filepath.Join(dbpath, "db-shm"),
		filepath.Join(dbpath, "db-wal"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			utils.Indent(log.WithField("path", path).Warn, 3)("DB file not found")
			continue
		}
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "removing file failed")
		}
		utils.Indent(log.WithField("path", path).Info, 3)("removed old DB file")
	}

	return restartDock()
}

func getiCloudDrivePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, "Library/Mobile Documents/com~apple~CloudDocs"), nil
}

func split[T any](buf []T, lim int) [][]T {
	var chunk []T
	chunks := make([][]T, 0, lim)
	for _, b := range buf {
		chunk = append(chunk, b)
		if len(chunk) == lim {
			chunks = append(chunks, chunk)
			chunk = nil
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}
	return chunks
}
