require 'mina/git'

set :branch, 'main'
set :execution_mode, 'system'
set :os, ['darwin']
set :arch, ['amd64']


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
    fetch(:os).each do |os|
      fetch(:arch).each do |arch|
        puts "build #{os}, #{arch} in #{fetch(:current_path)}"
      end
    end
    # dist_name = "rye-#{version}-macos-arm64"
    # dist_dir = "./dist/#{dist_name}"
    # sh %(BUILD=$(eval git rev-parse HEAD) && go build -o #{dist_dir}/rye -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=#{version}" ./cmd/cli/...)
    # sh %(touch #{dist_dir}/LICENSE)
    # sh %(touch #{dist_dir}/README.md)

    # # build a tarball
    # in_path("./dist/rye-#{version}-macos-arm64/README.md") do
    #   command %(tar czf ./dist/rye-#{version}-macos-arm64.tar.gz ./dist/rye-#{version}-macos-arm64)
    # end


    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/server)
    # sh %(GOOS=linux GOARCH=amd64 go build -o ./dist/linux/amd64/ ./cmd/cli)
  end
end