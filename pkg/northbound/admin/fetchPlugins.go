package admin

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
	"text/template/parse"
	"time"

	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/onosproject/onos-api/go/onos/config/admin"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const pluginListTemplateVerbose = "table{{.Id}}\t{{.Status}}\t{{.Endpoint}}\t{{.Info.Name}}\t{{.Info.Version}}\t{{.Error}}\t{{.Info.ModelData}}"

// Format defines a type for a string that can be used as template to format data.
type Format string

var nameFinder = regexp.MustCompile(`\.([\._A-Za-z0-9]*)}}`)
var outputWriter io.Writer

func ListPlugins() error {
	// opts := []grpc.DialOption{
	// 	grpc.WithInsecure(),
	// }
	var opts []grpc.DialOption
	//opts := credentials.ClientCredentials("onos-config")
	cert, err := tls.X509KeyPair([]byte(certs.DefaultClientCrt), []byte(certs.DefaultClientKey))
	if err != nil {
		fmt.Println("custom: error getting crts")
		return err
	}
	opts = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})),
	}

	conn, err := grpc.Dial("onos-config:5150", opts...)
	if err != nil {
		return err
	} else {
		fmt.Println("Connection succesfull")
	}
	client := admin.CreateConfigAdminServiceClient(conn)

	req := admin.ListModelsRequest{
		Verbose: true,
	}

	stream, err := client.ListRegisteredModels(context.Background(), &req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println("List plugins succesfull")
	}
	//fmt.Println(stream)
	var tableFormat Format
	tableFormat = pluginListTemplateVerbose

	allPlugins := []*admin.ModelPlugin{}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			if e := tableFormat.Execute(outputWriter, false, 0, allPlugins); e != nil {
				return e
			}
			return nil
		}
		if err != nil {
			return err
		}
		allPlugins = append(allPlugins, in)
	}
	fmt.Println("All plugins= ", allPlugins)
	fmt.Println("RW paths= ", allPlugins[0].GetInfo().GetReadWritePath())
	return nil
}

// Execute compiles the template and prints the output
func (f Format) Execute(writer io.Writer, withHeaders bool, nameLimit int, data interface{}) error {
	var tabWriter *tabwriter.Writer
	format := f

	if f.IsTable() {
		tabWriter = tabwriter.NewWriter(writer, 0, 4, 4, ' ', 0)
		format = Format(strings.TrimPrefix(string(f), "table"))
	}

	funcmap := template.FuncMap{
		"timestamp": formatTimestamp,
		"since":     formatSince,
		"gosince":   formatGoSince}

	tmpl, err := template.New("output").Funcs(funcmap).Parse(string(format))
	if err != nil {
		return err
	}

	if f.IsTable() && withHeaders {
		header := GetHeaderString(tmpl, nameLimit)

		if _, err = tabWriter.Write([]byte(header)); err != nil {
			return err
		}
		if _, err = tabWriter.Write([]byte("\n")); err != nil {
			return err
		}

		slice := reflect.ValueOf(data)
		if slice.Kind() == reflect.Slice {
			for i := 0; i < slice.Len(); i++ {
				fmt.Println(slice.Index(i))
				if err = tmpl.Execute(tabWriter, slice.Index(i).Interface()); err != nil {
					return err
				}
				if _, err = tabWriter.Write([]byte("\n")); err != nil {
					return err
				}
			}
		} else {
			if err = tmpl.Execute(tabWriter, data); err != nil {
				return err
			}
			if _, err = tabWriter.Write([]byte("\n")); err != nil {
				return err
			}
		}
		tabWriter.Flush()
		return nil
	}

	slice := reflect.ValueOf(data)
	if slice.Kind() == reflect.Slice {
		for i := 0; i < slice.Len(); i++ {
			if err = tmpl.Execute(writer, slice.Index(i).Interface()); err != nil {
				return err
			}
			if _, err = writer.Write([]byte("\n")); err != nil {
				return err
			}
		}
	} else {
		if err = tmpl.Execute(writer, data); err != nil {
			return err
		}
		if _, err = writer.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil

}

// IsTable returns a bool if the template is a table
func (f Format) IsTable() bool {
	return strings.HasPrefix(string(f), "table")
}

// GetHeaderString extract the set of column names from a template.
func GetHeaderString(tmpl *template.Template, nameLimit int) string {
	var header string
	for _, n := range tmpl.Tree.Root.Nodes {
		switch n.Type() {
		case parse.NodeText:
			header += n.String()
		case parse.NodeString:
			header += n.String()
		case parse.NodeAction:
			found := nameFinder.FindStringSubmatch(n.String())
			if len(found) == 2 {
				if nameLimit > 0 {
					parts := strings.Split(found[1], ".")
					start := len(parts) - nameLimit
					if start < 0 {
						start = 0
					}
					header += strings.ToUpper(strings.Join(parts[start:], "."))
				} else {
					header += strings.ToUpper(found[1])
				}
			}
		}
	}
	return header
}

//////////// Time functions /////////////////

// formats a Timestamp proto as a RFC3339 date string
func formatTimestamp(tsproto *timestamppb.Timestamp) (string, error) {
	if tsproto == nil {
		return "", nil
	}
	return tsproto.AsTime().Truncate(time.Second).Format(time.RFC3339), nil
}

// Computes the age of a timestamp and returns it in HMS format
func formatGoSince(ts time.Time) (string, error) {
	return time.Since(ts).Truncate(time.Second).String(), nil
}

// Computes the age of a timestamp and returns it in HMS format
func formatSince(tsproto *timestamppb.Timestamp) (string, error) {
	if tsproto == nil {
		return "", nil
	}
	return time.Since(tsproto.AsTime()).Truncate(time.Second).String(), nil
}
