package exporter

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/tobias-urdin/snapback/internal/message"
)

func buildImageSpec(req *message.ExportRequestV1) string {
	if req.Snapshot == "" {
		return fmt.Sprintf("%s/%s", req.Pool, req.Image)
	}

	return fmt.Sprintf("%s/%s@%s", req.Pool, req.Image, req.Snapshot)
}

func exportDiff(req *message.ExportRequestV1, w io.Writer) error {
	imageSpec := buildImageSpec(req)

	exportCmd := exec.Command("/bin/rbd", "export-diff", imageSpec, "-")
	fmt.Printf("export cmd: %v\n", exportCmd)

	out, err := exportCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := exportCmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(w, out); err != nil {
		return err
	}

	if err := exportCmd.Wait(); err != nil {
		return err
	}

	return nil
}
