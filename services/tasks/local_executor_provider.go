package tasks

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
)

// LocalExecutorProvider is the ExecutorProvider for the "local" strategy: tasks run
// as subprocesses on the runner host. The only shared dependency is the AccessKey
// installer, which materializes SSH keys / vault password files onto the local
// filesystem before each task — wiring it on the Provider keeps per-task construction
// cheap (no re-discovery of the installer on every dispatch).
type LocalExecutorProvider struct {
	keyInstaller db_lib.AccessKeyInstaller

	// repoLock serializes git operations on shared repository directories
	// across parallel tasks of the same template running on this runner.
	// Intentionally separate from TaskPool.repoLock: LocalExecutorProvider
	// (runner process) and TaskPool (server process) never operate on the
	// same directories, so they need no shared lock.
	repoLock *KeyLock
}

// NewLocalExecutorProvider takes the AccessKey installer the runner constructed at
// startup. The Provider does not own the installer's lifecycle — it just references
// it for every per-task Executor it produces.
func NewLocalExecutorProvider(keyInstaller db_lib.AccessKeyInstaller) *LocalExecutorProvider {
	return &LocalExecutorProvider{
		keyInstaller: keyInstaller,
		repoLock:     &KeyLock{},
	}
}

// NewExecutor returns a freshly-wired *LocalExecutor. db_lib.CreateApp is called
// here rather than inside LocalExecutor.Prepare so the executor arrives with a
// non-nil App — Prepare's contract is "do the I/O", not "build the structure".
func (p *LocalExecutorProvider) NewExecutor(task db.Task, template db.Template, inventory db.Inventory, repository db.Repository, environment db.Environment, jwt string) (Executor, error) {
	repository.WorkingCopyPath = resolveTaskCopyPath(repository, template, task)

	return &LocalExecutor{
		Task:        task,
		Template:    template,
		Inventory:   inventory,
		Repository:  repository,
		Environment: environment,
		// Survey secret variables delivered by the server in the job payload
		// (see JobData in the runner API).
		Secret:       task.Secret,
		KeyInstaller: p.keyInstaller,
		App:          db_lib.CreateApp(template, repository, inventory, nil),
		JWT:          jwt,
		RepoLock:     p.repoLock,
	}, nil
}
