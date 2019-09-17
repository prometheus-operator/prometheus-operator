task :deps do
  sh 'go get github.com/alecthomas/gometalinter'
  sh 'gometalinter --install'
  sh 'go get -d -t ./...'
  sh 'gem install rake rubocop'
  sh 'bundler install'
end

task :build do
  sh 'CGO_ENABLED=0 go build -o liche'
end

task :unit_test do
  sh 'go test -covermode atomic -coverprofile coverage.txt'
end

task integration_test: :build do
  sh 'bundler exec cucumber PATH=$PWD:$PATH'
end

task test: %i[unit_test integration_test]

task :format do
  sh 'go fix ./...'
  sh 'go fmt ./...'

  Dir.glob '**/*.go' do |file|
    sh "goimports -w #{file}"
  end

  sh 'rubocop -a'
end

task :lint do
  sh 'gometalinter ./...'
  sh 'rubocop'
end

task install: %i[deps test build] do
  sh 'go get ./...'
end

task default: %i[test build]

task :clean do
  sh 'git clean -dfx'
end
