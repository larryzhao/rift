require 'mina/git'

set :branch, 'main'
set :execution_mode, 'system'

desc 'Release a new version'
task :release do
  run(:local) do
    version = ENV['version']

    if version.nil? || version.empty?
      abort "err: version not specified"
    end

    Gem::Version.new(version)

    # check if git now on main and 
    sh %{
      if [[ $(git branch --show-current) != "add-build-action" ]]; then
        echo "err: branch is not main"
        exit 1
      fi
    }

    # check if there's no unstaged changes
    sh %{
      if [[ `git status --porcelain` ]]; then
        echo "err unstaged changes found"
        exit 1
      fi
    }

    # # create tag
    # sh %(git tag #{version})
      


    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/server)
    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/cli)
  end
end