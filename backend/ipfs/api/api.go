package api

import (
	"context"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/lib/rest"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	srv *rest.Client // the connection to the server
}

func NewApi(client *http.Client, endpoint string) *Client {
	api := Client{
		srv: rest.NewClient(client).SetRoot(endpoint),
	}
	api.srv.SetErrorHandler(errorHandler)
	return &api
}

// errorHandler parses a non 2xx error response into an error
func errorHandler(resp *http.Response) error {
	// Decode error response
	errResponse := new(Error)
	err := rest.DecodeJSON(resp, &errResponse)
	if err != nil {
		fs.Debugf(nil, "Couldn't decode error response: %v", err)
	}
	return errResponse
}

// Add file to IPFS
// /api/v0/add
func (a *Client) Add(ctx context.Context, in io.Reader, name string, options ...fs.OpenOption) (result *FileAdded, err error) {
	opts := rest.Opts{
		Method:               "POST",
		Path:                 "/api/v0/add",
		MultipartParams:      url.Values{},
		MultipartContentName: "file",
		MultipartFileName:    name,
		Body:                 in,
		Parameters: url.Values{
			"pin": []string{"false"},
		},
		Options: options,
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Check that the endpoint is read only
func (a *Client) IsReadOnly(ctx context.Context) (isReadOnly bool, err error) {
	opts := rest.Opts{
		Method: "POST",
		Path:   "/api/v0/add",
	}
	resp, err := a.srv.Call(ctx, &opts)
	if resp != nil && resp.StatusCode == 404 {
		// 404 => endpoint is read only
		return true, nil
	}

	if resp != nil && resp.StatusCode == 400 {
		// 400 => endpoint is read/write
		return false, nil
	}

	return false, err
}

// List file in IPFS path
// /api/v0/ls
func (a *Client) Ls(ctx context.Context, path string) ([]Link, error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/ls",
		Parameters: url.Values{
			"arg": []string{path},
		},
	}
	var result List
	_, err := a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	// Only one path provided so we get the first object links
	return result.Objects[0].Links, nil
}

// Read file in IPFS path
// /api/v0/cat
func (a *Client) Cat(ctx context.Context, objectPath string, objectSize int64, options ...fs.OpenOption) (result io.ReadCloser, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/cat",
		Parameters: url.Values{
			"arg": []string{objectPath},
		},
		Options: options,
	}

	for _, option := range options {
		seekOption, isSeek := option.(*fs.SeekOption)
		if isSeek {
			offset := strconv.FormatInt(seekOption.Offset, 10)
			opts.Parameters.Add("offset", offset)
		}
		rangeOption, isRange := option.(*fs.RangeOption)
		if isRange {
			if rangeOption.Start < 0 {
				offset := strconv.FormatInt(objectSize-rangeOption.End, 10)
				opts.Parameters.Add("offset", offset)
			} else {
				offset := strconv.FormatInt(rangeOption.Start, 10)
				opts.Parameters.Add("offset", offset)

				if rangeOption.End > rangeOption.Start {
					length := strconv.FormatInt(rangeOption.End-rangeOption.Start+1, 10)
					opts.Parameters.Add("length", length)
				}
			}
		}
	}
	resp, err := a.srv.Call(ctx, &opts)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Get IPFS DAG object stat
// /api/v0/object/stat
func (a *Client) ObjectStat(ctx context.Context, objectPath string) (result *ObjectStat, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/object/stat",
		Parameters: url.Values{
			"arg": []string{objectPath},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Patch a IPFS DAG object by adding (or replacing) a link.
// /api/v0/object/patch/add-link
func (a *Client) ObjectPatchAddLink(ctx context.Context, rootHash string, path string, linkHash string) (result *HasHash, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/object/patch/add-link",
		Parameters: url.Values{
			"arg": []string{
				rootHash, path, linkHash,
			},
			"create": []string{"true"},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Patch a IPFS DAG object by removing a link.
// /api/v0/object/patch/rm-link
func (a *Client) ObjectPatchRmLink(ctx context.Context, rootHash string, path string) (result *HasHash, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/object/patch/rm-link",
		Parameters: url.Values{
			"arg": []string{
				rootHash, path,
			},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create a new empty dir IPFS DAG object
// /api/v0/object/new
func (a *Client) ObjectNewDir(ctx context.Context) (result *HasHash, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/object/new",
		Parameters: url.Values{
			"arg": []string{"unixfs-dir"},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Diff two IPFS DAG object
// /api/v0/object/diff
func (a *Client) ObjectDiff(ctx context.Context, object1 string, object2 string) (result *ObjectDiff, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/object/diff",
		Parameters: url.Values{
			"arg": []string{object1, object2},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get file stat in IPFS MFS
// /api/v0/files/stat
func (a *Client) FilesStat(ctx context.Context, file string) (result *HasHash, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/files/stat",
		Parameters: url.Values{
			"arg": []string{file},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Copy IPFS file to IPFS MFS
// /api/v0/files/cp
func (a *Client) FilesCp(ctx context.Context, from string, to string) error {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/files/cp",
		Parameters: url.Values{
			"arg": []string{from, to},
		},
	}
	_, err := a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}

// Remove a IPFS MFS file
// /api/v0/files/rm
func (a *Client) FilesRm(ctx context.Context, dir string) error {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/files/rm",
		Parameters: url.Values{
			"arg":       []string{dir},
			"recursive": []string{"true"},
		},
	}
	_, err := a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}

// IPFS repo garbage collecting
// /api/v0/repo/gc
func (a *Client) RepoGc(ctx context.Context) error {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/repo/gc",
	}
	_, err := a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}

// Resolve an IPNS path to IPFS path
// /api/v0/name/resolve
func (a *Client) NameResolve(ctx context.Context, ipnsPath string) (result *HasPath, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/name/resolve",
		Parameters: url.Values{
			"arg": []string{ipnsPath},
		},
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// List IPNS keys
// /api/v0/key/list
func (a *Client) KeyList(ctx context.Context) (result *KeyList, err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/key/list",
	}
	_, err = a.srv.CallJSON(ctx, &opts, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Publish a IPNS
// /api/v0/name/publish
func (a *Client) NamePublish(ctx context.Context, ipfsPath string, key string) (err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/name/publish",
		Parameters: url.Values{
			"arg": []string{ipfsPath},
			"key": []string{key},
		},
	}
	_, err = a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}

// Pin an IPFS hash
// /api/v0/pin/add
func (a *Client) PinAdd(ctx context.Context, hash string, recursive bool) (err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/pin/add",
		Parameters: url.Values{
			"arg":       []string{hash},
			"recursive": []string{strconv.FormatBool(recursive)},
		},
	}
	_, err = a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}

// Unpin an IPFS hash
// /api/v0/pin/rm
func (a *Client) PinRm(ctx context.Context, hash string) (err error) {
	opts := rest.Opts{
		Method: "GET",
		Path:   "/api/v0/pin/rm",
		Parameters: url.Values{
			"arg": []string{hash},
		},
	}
	_, err = a.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	return nil
}
