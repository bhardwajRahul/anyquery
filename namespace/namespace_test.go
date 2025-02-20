package namespace

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/julien040/anyquery/rpc"
	"github.com/stretchr/testify/require"
)

func TestNamespace(t *testing.T) {
	os.Mkdir("_test", 0777)
	passed := t.Run("All three test plugins build correctly", func(t *testing.T) {
		err := exec.Command("go", "build", "-o", "_test/normalplugin.out", "../test/normalplugin.go").Run()
		require.NoError(t, err, "The normal plugin should build correctly")

		err = exec.Command("go", "build", "--tags", "vtable", "-o", "_test/rawplugin.out", "../test/rawplugin.go").Run()
		require.NoError(t, err, "The raw plugin should build correctly")

		err = exec.Command("go", "build", "--tags", "vtable", "-o", "_test/rawplugin2.out", "../test/rawplugin2.go").Run()
		require.NoError(t, err, "The lib plugin should build correctly")
	})

	if !passed {
		t.Log("Can't build the plugins, skipping the tests")
		return
	}

	var namespace *Namespace

	t.Run("It's possible to init a namespace", func(t *testing.T) {
		namespace = &Namespace{}
		err := namespace.Init(NamespaceConfig{
			InMemory:           true,
			PageCacheSize:      50000,
			EnforceForeignKeys: true,
		})
		require.NoError(t, err, "The namespace should be initialized")
	})

	// Test the connection string
	t.Run("The connection string is set correctly for in-memory DB", func(t *testing.T) {
		require.Equal(t, "file:anyquery.db?_cache_size=-50000&mode=memory&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON",
			namespace.connectionString, "The connection string should be correct")

	})

	t.Run("The connection string is set correctly for a file DB", func(t *testing.T) {
		var err error
		namespace, err = NewNamespace(NamespaceConfig{
			Path:          "test.db",
			PageCacheSize: 1000,
		})
		require.NoError(t, err, "The namespace should be initialized")
		require.Equal(t, "file:test.db?_cache_size=-1000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=OFF",
			namespace.connectionString, "The connection string should be correct")
	})
	t.Run("The connection string is set correctly for a file DB with a custom connection string", func(t *testing.T) {
		var err error
		namespace, err = NewNamespace(NamespaceConfig{
			ConnectionString: "file:mytest.db?cache=shared&_foreign_keys=OFF",
			InMemory:         true,
			Path:             "random garbage that doesn't matter",
		})
		require.NoError(t, err, "The namespace should be initialized")
		require.Equal(t, "file:mytest.db?cache=shared&_foreign_keys=OFF",
			namespace.connectionString, "The connection string should be correct")
	})

	t.Run("The connection string is set correctly for a read-only file DB", func(t *testing.T) {
		var err error
		namespace, err = NewNamespace(NamespaceConfig{
			ReadOnly: true,
		})
		require.NoError(t, err, "The namespace should be initialized")
		require.Equal(t, "file:anyquery.db?_cache_size=-50000&mode=ro&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=OFF",
			namespace.connectionString, "The connection string should be correct")
	})

	t.Run("The read only flag is ignored for in-memory DB", func(t *testing.T) {
		var err error
		namespace, err = NewNamespace(NamespaceConfig{
			ReadOnly: true,
			InMemory: true,
		})
		require.NoError(t, err, "The namespace should be initialized")
		require.Equal(t, "file:anyquery.db?_cache_size=-50000&mode=memory&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=OFF",
			namespace.connectionString, "The connection string should be correct")
	})

	t.Run("The GetConnectionString method works", func(t *testing.T) {
		require.Equal(t, "file:anyquery.db?_cache_size=-50000&mode=memory&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=OFF",
			namespace.GetConnectionString(), "The connection string should be correct")
		require.Equal(t, namespace.connectionString, namespace.GetConnectionString(), "The connection string should be correct")
	})

	// Test that loading a shared library works
	// To do so, we will use SQLean as a shared library
	t.Run("Load a shared library", func(t *testing.T) {
		err := downloadSQLean()
		if err != nil {
			t.Log("Can't download SQLean, skipping the test")
			t.Log(err)
			t.Skip()
		}

		// Create a new namespace
		namespace, err := NewNamespace(NamespaceConfig{
			InMemory: true,
		})
		require.NoError(t, err, "The namespace should be initialized")

		// Load the shared library
		// SQLite will automatically add the extension (dll, so, dylib) to the name
		// See https://tinyurl.com/2d2zfrfe
		err = namespace.LoadSharedExtension("_test/sqlean", "")

		require.NoError(t, err, "The shared library should load correctly")

		// Register the connection and run a simple query
		db, err := namespace.Register("") // We left a blank string to let the namespace choose the name

		require.NoError(t, err, "The connection should be registered")

		// Run a simple query
		rows, err := db.Query("SELECT text_left('Hello world!', 5)")
		require.NoError(t, err, "The query should work")

		// Check the result
		require.True(t, rows.Next(), "There should be a row")
		var result string
		err = rows.Scan(&result)
		require.NoError(t, err, "The scan should work")
		require.Equal(t, "Hello", result, "The result should be correct")

		// Close the rows
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Close the database
		err = db.Close()
		require.NoError(t, err, "The database should be closed")
	})

	// Test that loading a Go plugin works
	/* t.Run("Load a Go plugin", func(t *testing.T) {
		// Create a new namespace
		namespace, err := NewNamespace(NamespaceConfig{
			InMemory: true,
		})
		require.NoError(t, err, "The namespace should be initialized")

		// Load the Go plugin
		err = namespace.LoadAnyqueryPlugin("_test/normalplugin.out", rpc.PluginManifest{
			Tables: []string{"test", "test2"},
		}, nil, 0)
		require.NoError(t, err, "The Go plugin should load correctly")

		// Register the connection and run a simple query
		db, err := namespace.Register("") // We left a blank string to let the namespace choose the name

		require.NoError(t, err, "The connection should be registered")

		// Run a simple query
		rows, err := db.Query("SELECT A.id, A.name FROM test A")
		require.NoError(t, err, "The query should work")

		// The plugin should return two rows
		i := 0
		for rows.Next() {
			i++
		}

		require.NoError(t, rows.Err(), "The rows should not have an error")

		require.Equal(t, 2, i, "The number of rows should be correct")

		// Close the rows
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Close the database
		err = db.Close()
		require.NoError(t, err, "The database should be closed")

	}) */

	t.Run("Ensure that user errors are handled correctly", func(t *testing.T) {
		namespace, err := NewNamespace(NamespaceConfig{
			InMemory: true,
		})
		require.NoError(t, err, "The namespace should be initialized")

		_, err = namespace.Register("mydb")
		require.NoError(t, err, "The connection should be registered")

		// Test that the path of a shared extension cannot be empty
		err = namespace.LoadSharedExtension("", "")
		require.Error(t, err, "The shared library should not load correctly if the path is empty")

		err = namespace.LoadAnyqueryPlugin("_test/normalplugin.out", rpc.PluginManifest{
			Tables: []string{"test", "test2"},
		}, nil, 0)
		require.Error(t, err, "The Go plugin should not load correctly if the namespace is already registered")

		err = namespace.LoadAnyqueryPlugin("", rpc.PluginManifest{}, nil, 0)
		require.Error(t, err, "The Go plugin should not load correctly if the path is empty")

		err = namespace.LoadSharedExtension("_test/sqlean", "")
		require.Error(t, err, "The shared library should not load correctly if the namespace is already registered")

		err = namespace.LoadGoPlugin(nil, "")
		require.Error(t, err, "The Go plugin should not load correctly if the namespace is already registered")

		_, err = namespace.Register("mydb2")
		require.Error(t, err, "The connection should not be registered if the namespace is already registered")

		// Ensure that if a namespace is already registered with a name, another namespace with the same name cannot be registered
		namespace2, err := NewNamespace(NamespaceConfig{
			InMemory: true,
		})
		require.NoError(t, err, "The namespace should be initialized")

		_, err = namespace2.Register("mydb")
		require.Error(t, err, "The connection should not be registered if the namespace is already registered")

		// Ensure that a namespace non-inited cannot register a connection
		namespace3 := &Namespace{}
		_, err = namespace3.Register("mydb")
		require.Error(t, err, "The connection should not be registered if the namespace is not initialized")

	})

	t.Run("Test dev functions", func(t *testing.T) {
		namespace, err := NewNamespace(NamespaceConfig{
			InMemory: true,
			DevMode:  true,
		})
		require.NoError(t, err, "The namespace should be initialized")

		// Register the namespace
		db, err := namespace.Register("mydb2")
		require.NoError(t, err, "The connection should be registered")
		defer db.Close()

		// Save the dev manifest to disk
		err = os.WriteFile("_test/devManifest.json", []byte(devManifestTesting), 0777)
		require.NoError(t, err, "The dev manifest should be saved")
		defer os.Remove("_test/devManifest.json")

		// Run a dev function
		rows, err := db.Query("SELECT load_dev_plugin(\"test1\", \"_test/devManifest.json\")")
		require.NoError(t, err, "The query should work")

		// Check the result
		require.True(t, rows.Next(), "There should be a row")
		var result string
		err = rows.Scan(&result)
		require.True(t, strings.Contains(result, "Successfully"), "The result should be correct")
		require.NoError(t, err, "The scan should work")
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Do queries on the plugin
		rows, err = db.Query("SELECT A.id, A.name FROM test1_test A")
		require.NoError(t, err, "The query should work")

		// Check the result
		i := 0
		for rows.Next() {
			i++
		}
		require.Equal(t, 2, i, "The number of rows should be correct")

		// Close the rows
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Unload the plugin
		rows, err = db.Query("SELECT unload_dev_plugin(\"test1\")")
		require.NoError(t, err, "The query should work")
		for rows.Next() {
		} // We have to call Next, otherwise the SELECT statement is not executed
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Test that we can't query the table anymore
		_, err = db.Query("SELECT A.id, A.name FROM test1_test A")
		require.Error(t, err, "The query should not work")

		// Test that we can load the plugin again
		rows, err = db.Query("SELECT load_dev_plugin(\"test1\", \"_test/devManifest.json\")")
		require.NoError(t, err, "The query should work")
		for rows.Next() {
		} // We have to call Next, otherwise the SELECT statement is not executed
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Query the plugin again
		rows, err = db.Query("SELECT A.id, A.name FROM test1_test A")
		require.NoError(t, err, "The query should work")
		// Check the result
		i = 0
		for rows.Next() {
			i++
		}
		require.Equal(t, 2, i, "The number of rows should be correct")
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

		// Reload the plugin
		rows, err = db.Query("SELECT reload_dev_plugin(\"test1\")")
		require.NoError(t, err, "The query should work")
		rows.Close()

		// Query the plugin again
		rows, err = db.Query("SELECT A.id, A.name FROM test1_test A")
		require.NoError(t, err, "The query should work")
		// Check the result
		i = 0
		for rows.Next() {
			i++
		}
		require.Equal(t, 2, i, "The number of rows should be correct")
		err = rows.Close()
		require.NoError(t, err, "The rows should be closed")

	})

}

var devManifestTesting string = `
{
	"executable": "_test/normalplugin.out",
	"tables": [
		"test",
		"test2"
	],
	"user_config": {
		"default": {
			"notInteresting": "value"
		}
	}
}
`

func downloadSQLean() error {
	// Find the right URL according to the OS and architecture
	urlToDownload := ""
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		urlToDownload = "https://github.com/nalgeon/sqlean/releases/download/0.22.0/sqlean-macos-arm64.zip"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "amd64" {
		urlToDownload = "https://github.com/nalgeon/sqlean/releases/download/0.22.0/sqlean-macos-x86.zip"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		urlToDownload = "https://github.com/nalgeon/sqlean/releases/download/0.22.0/sqlean-linux-x86.zip"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		urlToDownload = "https://github.com/nalgeon/sqlean/releases/download/0.22.0/sqlean-linux-arm64.zip"
	} else if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		urlToDownload = "https://github.com/nalgeon/sqlean/releases/download/0.22.0/sqlean-win-x64.zip"
	} else {
		return errors.New("unsupported OS or architecture to download SQLean")
	}

	// Run curl to download the file
	err := exec.Command("curl", "-C", "-", "-L", "-o", "_test/sqlean.out", urlToDownload).Run()
	if err != nil {
		return err
	}

	// Unzip the file
	// We use -o to overwrite the files
	// We use -d to specify the directory
	return exec.Command("unzip", "-o", "_test/sqlean.out", "-d", "_test").Run()

}
