package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

func main() {
	uuidGenerator := ec2.NewUUIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)

	keyPairCreator := ec2.NewKeyPairCreator(uuidGenerator.Generate)
	keyPairChecker := ec2.NewKeyPairChecker()
	keyPairManager := ec2.NewKeyPairManager(keyPairCreator, keyPairChecker, logger)
	keyPairSynchronizer := unsupported.NewKeyPairSynchronizer(keyPairManager)
	stateStore := storage.NewStore()
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(logger)
	infrastructureCreator := unsupported.NewInfrastructureCreator(templateBuilder, stackManager)
	awsClientProvider := aws.NewClientProvider()
	sslKeyPairGenerator := ssl.NewKeyPairGenerator(time.Now, rsa.GenerateKey, x509.CreateCertificate)
	boshInitManifestBuilder := boshinit.NewManifestBuilder(logger, sslKeyPairGenerator)
	boshDeployer := unsupported.NewBOSHDeployer(boshInitManifestBuilder)

	app := application.New(application.CommandSet{
		"help":    commands.NewUsage(os.Stdout),
		"version": commands.NewVersion(os.Stdout),
		"unsupported-deploy-bosh-on-aws-for-concourse": unsupported.NewDeployBOSHOnAWSForConcourse(infrastructureCreator, keyPairSynchronizer, awsClientProvider, boshDeployer),
	}, stateStore, commands.NewUsage(os.Stdout).Print)

	err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
		os.Exit(1)
	}
}
