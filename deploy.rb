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

    # build binaries
    sh %(BUILD=$(eval git rev-parse HEAD) && go build -o ./dist/rye-#{version}-macos-arm64/rye -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=#{version}" ./cmd/cli/...)
    # sh %(git tag #{version})


    # 

    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/server)
    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/cli)
  end
end