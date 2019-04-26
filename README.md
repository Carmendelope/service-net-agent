# service-net-agent
Service Net Agent

## Plugin infrastructure

Agent plugins are defined in `internal/pkg/plugin`, in a subdirectory per plugin. A plugin can be instantiated, started and stopped. Furthermore, each plugin has a set of commands with parameters that it can execute.

A plugin should register itself with the (default) plugin registry in its `init()` function by calling `plugin.Register(name PluginName, newFunc NewPluginFunc)`. `NewPluginFunc` is the constructor function that will return an instance of the specific plugin - which is an implementation of the `Plugin` interface:


```Golang
type Plugin interface {
        StartPlugin() (derrors.Error)
        StopPlugin()

        // Retrieve the function for a specific command to be executed in the
        // caller. Alternative would be an ExecuteCommand() function on
        // the plugin, but that would duplicate implementation for each plugin
        // that we can now implement in the registry.
        GetCommand(cmd CommandName) CommandFunc
}
```

Internally, the plugin has a mapping from command names (`type CommandName string`) to command execution functions (`type CommandFunc func(params map[string]string) (string, derrors.Error)`), which could be static or dynamic. The `GetCommand` function returns the appropriate execution function for a command.

An example plugin is provided in `internal/pkg/plugin/ping`.

**NOTE:** Make sure your plugin is enabled by importing it in `internal/app/run/plugins.go`.

### Workflow

Skipping over details of an agent receiving commands and scheduling those, the actual plugin workflow is as follows (from the start of the agent):

1. Default registry is created in `registry.go:init()`
2. Each plugin calls `plugin.Register()` from its `init()` (which operates on the default registry)
3. An operation that requests to start a plugin is received
 - The parameters of the operation are converted to a Viper configuration;
 - The default registry function `plugin.StartPlugin(name PluginName, conf *viper.Viper)` is called;
 - The registry creates an instance of the plugin, passing the configuration
 - The registry calls `StartPlugin()` on the plugin to spawn any background processes that the plugin needs
4. An operation for a running plugin is received
 - The default registry function `plugin.ExecuteCommand(name PluginName, cmd CommandName, params map[string]string)` is called
 - The registry checks if the plugin is running
 - The registry gets the actual command function by calling `GetCommand(cmdName)` on the plugin; if `nil` is returned an error is returned to the caller
 - The registry calls the command function on the plugin

 Stop plugin functionality works similar to the start functionality above; receiving, scheduling and execution operations in a Goroutine is handled by the main thread of the agent and is captured in what is called "caller" above. A description of that is outside the scope of the plugin architecture.